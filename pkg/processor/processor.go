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
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Instance profile key, watching
	WatcherPrefix = "/iam/instance-profile/"
	// Credential timer key, watching, ttl
	KeeperPrefix = "/iam/lease/"
	// Register key, watching, ttl
	RegPrefix = "/iam/reg/"

	CredPrefix = "/iam/credential/"
	// Lock key, ttl
	MutexPrefix = "/iam/mutex/"
	// Flag key, ttl
	FlagPrefix = "/iam/flag/"

	MaxQueueSize = 20
	LockTTL      = 5
)

type Processor interface {
	Process()
}

// Watcher monitoring the /iam/instance-profile/
// To keep tracking the update/remove for instance profile
type watcher struct {
	cli      *etcdv3.Client
	syncChan chan<- string
}

func NewWatcher(cli *etcdv3.Client) Processor {
	return &watcher{cli, DefaultChans.SyncChan}
}

func (wa watcher) Process() {
	go wa.cli.WatchPrefix(WatcherPrefix, func(t int32, k, v []byte) error {
		// Trigger the sync immediately
		wa.syncChan <- strings.TrimPrefix(string(k), WatcherPrefix)

		switch t {
		case 0:
			fmt.Println("Put event:", string(k), string(v))

		case 1:
			fmt.Println("Delete event:", string(k))
		}
		return nil
	})
}

// Keeper monitoring the /iam/lease/
// To keep tracking the lifecycle/timeout of temporary credential
type keeper struct {
	cli      *etcdv3.Client
	syncChan chan<- string
}

func NewKeeper(cli *etcdv3.Client) Processor {
	return &keeper{cli, DefaultChans.SyncChan}
}

func (ke keeper) Process() {
	go ke.cli.WatchPrefix(KeeperPrefix, func(t int32, k, v []byte) error {
		switch t {
		case 0:
			//fmt.Println("Put event:", string(k), string(v))
		case 1:
			//fmt.Println("Delete event:", string(k))
			ke.syncChan <- strings.TrimPrefix(string(k), KeeperPrefix)
		}
		return nil
	})
}

// Sync attempt to talk with sts
// Write the temporary credential to metadata server within /iam/credential/
type sync struct {
	etcdCli  *etcdv3.Client
	stsCli   *sts.Client
	syncChan <-chan string
	ttl      int64
	cluster  *Cluster
}

func NewSync(etcdCli *etcdv3.Client, stsCli *sts.Client, ttl int64, cluster *Cluster) Processor {
	return &sync{etcdCli, stsCli, DefaultChans.SyncChan, ttl, cluster}
}

func (sy sync) Process() {
	go func() {
		fmt.Println("Waiting on sync channel...")
		for v := range sy.syncChan {
			fmt.Println("SyncChan got an update on:", v)
			if sy.cluster.Mock {
				sy.doSyncMock(v)
			} else {
				sy.doSync(v)
			}
		}
	}()
}

func (sy sync) doSync(id string) {
	go func() {
		rid := "#" + fmt.Sprint(rand.Int63()%32768)
		s, m, err := sy.etcdCli.Lock(MutexPrefix+id, LockTTL)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer sy.etcdCli.Unlock(s, m)

		// Check the offset flag
		flag, err := sy.etcdCli.GetSingleValue(FlagPrefix + id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Skip if offset <= cluster size
		if flag != nil {
			offset, _ := strconv.ParseInt(string(flag), 0, 0)

			if offset < sy.cluster.Size {
				offset = offset + 1
				sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
				return
			}
		}

		// Get bundle from WatcherPrefix
		bundle, err := sy.etcdCli.GetSingleValue(WatcherPrefix + id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Bundle goes to nil which means the key does not exist or error occurred
		// We attempt to delete the credential key and left the lease for self-deleting
		if bundle == nil {
			sy.etcdCli.DeleteKey(CredPrefix + id)
			fmt.Println("Sync stoped due to missing key:", WatcherPrefix+id)
			return
		}
		// Attempt to call sts AssumeRole
		credential, err := sy.stsCli.AssumeRole(bundle)
		// Succeed
		if err == nil {
			// Set credential data to CredPrefix
			err = sy.etcdCli.SetSingleValue(CredPrefix+id, string(credential))
			if err != nil {
				fmt.Println(err)
				return
			}

			// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
			// The flag will be deleted after LockTTL
			err = sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			fmt.Println(err)
		}

		// Whatever, We keep setting ttl to KeeperPrefix
		err = sy.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, sy.ttl)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Sync successfully on id:", fmt.Sprintf("%v", sy.cluster.Pid), id)
	}()
}

func (sy sync) doSyncMock(id string) {
	go func() {
		rid := "#" + fmt.Sprint(rand.Int63()%32768)
		//fmt.Println("### Cluster size is", sy.cluster.Size)
		//fmt.Println("### Try to lock:", sy.cluster.Pid)
		s, m, err := sy.etcdCli.Lock(MutexPrefix+id, LockTTL)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer sy.etcdCli.Unlock(s, m)

		//fmt.Println("### Get the lock:", sy.cluster.Pid)
		// Check the offset flag
		flag, err := sy.etcdCli.GetSingleValue(FlagPrefix + id)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("### Flag is", string(flag))
		// Skip if offset <= cluster size
		if flag != nil {
			offset, _ := strconv.ParseInt(string(flag), 0, 0)

			//fmt.Println("### offset is ", offset)
			if offset < sy.cluster.Size {
				offset = offset + 1
				sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
				fmt.Println("### SKIP and set offset to", offset, id)
				return
			}
			fmt.Println("!!! skip is ignore due to offset is %v, cluster size is", offset, sy.cluster.Size)
		}

		// Get bundle from WatcherPrefix
		bundle, err := sy.etcdCli.GetSingleValue(WatcherPrefix + id)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("### Bundle is", bundle)
		// Bundle goes to nil which means the key does not exist or error occurred
		// We attempt to delete the credential key and left the lease for self-deleting
		if bundle == nil {
			sy.etcdCli.DeleteKey(CredPrefix + id)
			fmt.Println("Sync stoped due to missing key:", WatcherPrefix+id)
			return
		}
		// Call sts AssumeRole
		credential, err := sy.stsCli.AssumeRoleMock(bundle)
		// Succeed
		if err == nil {
			// Set credential data to CredPrefix
			err = sy.etcdCli.SetSingleValue(CredPrefix+id, string(credential)+fmt.Sprintf("%v", sy.cluster.Pid))
			if err != nil {
				fmt.Println(err)
				return
			}

			// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
			// The flag will be deleted after LockTTL
			err = sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			fmt.Println(err)
		}

		//fmt.Println("### credential write to", CredPrefix+id)
		// Set ttl to KeeperPrefix
		err = sy.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, 10)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("### flag set to %s value is", FlagPrefix + id, 1)
		fmt.Println("@@@ SYNC successfully on id:", fmt.Sprintf("%v", sy.cluster.Pid), id)
	}()
}

// Sync attempt to write the data to /iam/credential/
type register struct {
	cli      *etcdv3.Client
	cluster  *Cluster
	doneChan chan struct{}
}

func NewRegister(cli *etcdv3.Client, cluster *Cluster) Processor {
	return &register{cli, cluster, DefaultChans.DoneChan}
}

func (re register) Process() {
	c := make(chan struct{})
	go func() {
		defer close(c)
		re.cluster.Pid = os.Getpid()
		re.doRegister()

		re.cluster.Size, _ = re.cli.GetKeyNumber(RegPrefix)
	}()
	<-c

	go re.cli.WatchPrefix(RegPrefix, func(t int32, k, v []byte) error {
		if t == 1 {
			if string(k) == re.cluster.RegKey {
				re.doRegister()
				//fmt.Println("Cluster Regkey changes to:", re.cluster.RegKey)
			}
		}
		re.cluster.Size, _ = re.cli.GetKeyNumber(RegPrefix)
		fmt.Println("Cluster size changes to:", re.cluster.Size)
		return nil
	})
	fmt.Println("Cluster Regkey is:", re.cluster.RegKey)
	fmt.Println("Cluster size is:", re.cluster.Size)
}

func (re register) doRegister() {
	key := fmt.Sprintf("%s%x", RegPrefix, re.cluster.Pid)
	for {
		_, regkey, err := re.cli.Register(key, LockTTL)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Duration(1) * time.Second)
		} else {
			re.cluster.RegKey = regkey
			break
		}
	}
}

type Cluster struct {
	Size   int64
	Pid    int
	RegKey string
	Mock   bool
}

type Chans struct {
	DoneChan chan struct{}
	ErrChan  chan error
	SyncChan chan string
}

func NewChans() *Chans {
	return &Chans{make(chan struct{}), make(chan error, MaxQueueSize), make(chan string, MaxQueueSize)}
}

var DefaultChans = NewChans()
