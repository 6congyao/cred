/*
 * Copyright (c) 2018. LuCongyao <6congyao@gmail.com> .
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this work except in compliance with the License.
 * You may obtain a copy of the License in the LICENSE file, or at:
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package processor

import (
	"cred/pkg/backend/etcdv3"
	"cred/pkg/backend/sts"
	"cred/utils/logger"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// Dispatcher attempt to dispatch the tasks to workerpool
// To write the temporary credential to metadata server within /iam/credential/
type dispatcher struct {
	etcdCli *etcdv3.Client
	stsCli  *sts.Client

	ttl      int64
	cluster  *Cluster
	syncChan <-chan string

	wp *workerPool
}

func NewDispatcher(etcdCli *etcdv3.Client, stsCli *sts.Client, ttl int64, cluster *Cluster) Processor {
	wp := &workerPool{
		MaxWorkersCount: DefaultConcurrency,
		LogAllErrors:    true,
		Logger:          logger.Error,
	}

	return &dispatcher{etcdCli, stsCli, ttl, cluster, DefaultChans.SyncChan, wp}
}

func (di *dispatcher) Process() {
	di.wp.WorkerFunc = di.handler
	di.wp.Start()

	reg := NewRegister(di.etcdCli, di.cluster)
	reg.Process()

	watcher := NewWatcher(di.etcdCli, di.wp)
	watcher.Process()

	keeper := NewKeeper(di.etcdCli, di.wp)
	keeper.Process()

	go func() {
		for v := range di.syncChan {
			if !di.wp.Serve(v) {
				logger.Warn.Printf("The connection cannot be served because Server.Concurrency limit exceeded")
				// The current server reached concurrency limit,
				// so give other concurrently running servers a chance
				// accepting incoming connections on the same address.
				//
				// There is a hope other servers didn't reach their
				// concurrency limits yet :)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (di *dispatcher) handler(id string) error {
	logger.Info.Printf("SyncChan got an update on: %s", id)
	if di.cluster.Mock {
		di.doSyncMock(id)
	} else {
		di.doSync(id)
	}
	return nil
}

func (di *dispatcher) doSync(id string) {
	rid := "#" + fmt.Sprint(rand.Int63()%32768)
	resetLease := false
	var leaseTTL int64

	s, m, err := di.etcdCli.Lock(MutexPrefix+id, LockTTL)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	defer func() {
		if resetLease {
			if leaseTTL == 0 {
				leaseTTL = di.ttl
			}
			// Whatever, We keep setting ttl to KeeperPrefix
			err = di.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, leaseTTL)
			if err != nil {
				logger.Error.Print(err)
			}
		}

		di.etcdCli.Unlock(s, m)
	}()

	// Check the offset flag
	flag, err := di.etcdCli.GetSingleValue(FlagPrefix + id)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	// Skip if offset <= cluster size
	if flag != nil {
		offset, _ := strconv.ParseInt(string(flag), 0, 0)

		if offset < di.cluster.Size {
			offset = offset + 1
			di.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
			return
		}
	}
	// Get bundle from WatcherPrefix
	bundle, err := di.etcdCli.GetSingleValue(WatcherPrefix + id)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	// Bundle goes to nil which means the key does not exist or error occurred
	// We attempt to delete the credential key and left the lease for self-deleting
	if bundle == nil {
		di.etcdCli.DeleteKey(CredPrefix + id)
		logger.Info.Printf("Sync stoped due to missing key: %s%s", WatcherPrefix, id)
		return
	}
	// Attempt to call sts AssumeRole
	credential, err := di.stsCli.AssumeRole(bundle)

	// The lease should be reset here
	resetLease = true
	leaseTTL = parseTTL(bundle)

	// Succeed
	if err == nil {
		// Set credential data to CredPrefix
		err = di.etcdCli.SetSingleValue(CredPrefix+id, string(credential))
		if err != nil {
			logger.Error.Print(err)
			return
		}

		// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
		// The flag will be deleted after LockTTL
		err = di.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
		if err != nil {
			logger.Error.Print(err)
			return
		}
	} else {
		logger.Error.Print(err)
	}

	logger.Info.Printf("Sync successfully on id: %v %s", di.cluster.Pid, id)
}

func (di *dispatcher) doSyncMock(id string) {
	rid := "#" + fmt.Sprint(rand.Int63()%32768)
	resetLease := false
	var leaseTTL int64

	//fmt.Println("### Cluster size is", sy.cluster.Size)
	//fmt.Println("### Try to lock:", sy.cluster.Pid)
	s, m, err := di.etcdCli.Lock(MutexPrefix+id, LockTTL)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	defer func() {
		if resetLease {
			if leaseTTL == 0 {
				leaseTTL = di.ttl
			}
			// Whatever, We keep setting ttl to KeeperPrefix
			err = di.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, 10)
			if err != nil {
				logger.Error.Print(err)
			}
		}

		di.etcdCli.Unlock(s, m)
	}()

	//fmt.Println("### Get the lock:", sy.cluster.Pid)
	// Check the offset flag
	flag, err := di.etcdCli.GetSingleValue(FlagPrefix + id)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	//fmt.Println("### Flag is", string(flag))
	// Skip if offset <= cluster size
	if flag != nil {
		offset, _ := strconv.ParseInt(string(flag), 0, 0)

		//fmt.Println("### offset is ", offset)
		if offset < di.cluster.Size {
			offset = offset + 1
			di.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
			logger.Info.Printf("### SKIP and set offset to %d %s", offset, id)
			return
		}
		logger.Info.Printf("!!! skip is ignore due to offset is %v, cluster size is %d", offset, di.cluster.Size)
	}

	// Get bundle from WatcherPrefix
	bundle, err := di.etcdCli.GetSingleValue(WatcherPrefix + id)
	if err != nil {
		logger.Error.Print(err)
		return
	}
	//fmt.Println("### Bundle is", bundle)
	// Bundle goes to nil which means the key does not exist or error occurred
	// We attempt to delete the credential key and left the lease for self-deleting
	if bundle == nil {
		di.etcdCli.DeleteKey(CredPrefix + id)
		logger.Info.Printf("Sync stoped due to missing key: %s%s", WatcherPrefix, id)
		return
	}
	// Call sts AssumeRole
	credential, err := di.stsCli.AssumeRoleMock(bundle)
	// The lease should be reset here
	resetLease = true
	leaseTTL = parseTTL(bundle)
	// Succeed
	if err == nil {
		// Set credential data to CredPrefix
		err = di.etcdCli.SetSingleValue(CredPrefix+id, string(credential)+fmt.Sprintf("%v", rid))
		if err != nil {
			logger.Error.Print(err)
			return
		}

		// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
		// The flag will be deleted after LockTTL
		err = di.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
		if err != nil {
			logger.Error.Print(err)
			return
		}
	} else {
		logger.Error.Print(err)
	}
	//fmt.Println("### credential write to", CredPrefix+id)
	logger.Info.Printf("@@@ SYNC successfully on id:%v %s", di.cluster.Pid, id)
}

// todo: parse the ttl from bundle, return 0 if error
func parseTTL(data []byte) (int64) {
	return 0
}