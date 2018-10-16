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

	DefaultCredentialLeaseTTL = 3600

	// DefaultConcurrency is the maximum number of concurrent tasks
	// the Processor may serve by default
	DefaultConcurrency = 256 * 1024
)

type Processor interface {
	Process()
}

type Cluster struct {
	Size     int64
	Hostname string
	Pid      int
	RegKey   string
	Mock     bool
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

// Logger is used for logging formatted messages.
type Logger interface {
	// Printf must have the same semantics as log.Printf.
	Printf(format string, args ...interface{})
}

// Sync attempt to talk with sts
// Write the temporary credential to metadata server within /iam/credential/
//type worker struct {
//	etcdCli  *etcdv3.Client
//	stsCli   *sts.Client
//	syncChan <-chan string
//	ttl      int64
//	cluster  *Cluster
//}
//
//func NewWorker(etcdCli *etcdv3.Client, stsCli *sts.Client, ttl int64, cluster *Cluster) Processor {
//	return &worker{etcdCli, stsCli, DefaultChans.SyncChan, ttl, cluster}
//}
//
//func (sy worker) Process() {
//	go func() {
//		logger.Info.Print("Waiting on sync channel...")
//		for v := range sy.syncChan {
//			logger.Info.Printf("SyncChan got an update on: %s", v)
//			if sy.cluster.Mock {
//				sy.doSyncMock(v)
//			} else {
//				sy.doSync(v)
//			}
//		}
//	}()
//}
//
//func (sy worker) doSync(id string) {
//	go func() {
//		rid := "#" + fmt.Sprint(rand.Int63()%32768)
//		s, m, err := sy.etcdCli.Lock(MutexPrefix+id, LockTTL)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		defer sy.etcdCli.Unlock(s, m)
//
//		// Check the offset flag
//		flag, err := sy.etcdCli.GetSingleValue(FlagPrefix + id)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		// Skip if offset <= cluster size
//		if flag != nil {
//			offset, _ := strconv.ParseInt(string(flag), 0, 0)
//
//			if offset < sy.cluster.Size {
//				offset = offset + 1
//				sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
//				return
//			}
//		}
//
//		// Get bundle from WatcherPrefix
//		bundle, err := sy.etcdCli.GetSingleValue(WatcherPrefix + id)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		// Bundle goes to nil which means the key does not exist or error occurred
//		// We attempt to delete the credential key and left the lease for self-deleting
//		if bundle == nil {
//			sy.etcdCli.DeleteKey(CredPrefix + id)
//			logger.Info.Printf("Sync stoped due to missing key: %s%s", WatcherPrefix, id)
//			return
//		}
//		// Attempt to call sts AssumeRole
//		credential, err := sy.stsCli.AssumeRole(bundle)
//		// Succeed
//		if err == nil {
//			// Set credential data to CredPrefix
//			err = sy.etcdCli.SetSingleValue(CredPrefix+id, string(credential))
//			if err != nil {
//				logger.Error.Print(err)
//				return
//			}
//
//			// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
//			// The flag will be deleted after LockTTL
//			err = sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
//			if err != nil {
//				logger.Error.Print(err)
//				return
//			}
//		} else {
//			logger.Error.Print(err)
//		}
//
//		// Whatever, We keep setting ttl to KeeperPrefix
//		err = sy.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, sy.ttl)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//
//		logger.Info.Printf("Sync successfully on id: %v %s", sy.cluster.Pid, id)
//	}()
//}
//
//func (sy worker) doSyncMock(id string) {
//	go func() {
//		rid := "#" + fmt.Sprint(rand.Int63()%32768)
//		//fmt.Println("### Cluster size is", sy.cluster.Size)
//		//fmt.Println("### Try to lock:", sy.cluster.Pid)
//		s, m, err := sy.etcdCli.Lock(MutexPrefix+id, LockTTL)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		defer sy.etcdCli.Unlock(s, m)
//
//		//fmt.Println("### Get the lock:", sy.cluster.Pid)
//		// Check the offset flag
//		flag, err := sy.etcdCli.GetSingleValue(FlagPrefix + id)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		//fmt.Println("### Flag is", string(flag))
//		// Skip if offset <= cluster size
//		if flag != nil {
//			offset, _ := strconv.ParseInt(string(flag), 0, 0)
//
//			//fmt.Println("### offset is ", offset)
//			if offset < sy.cluster.Size {
//				offset = offset + 1
//				sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(offset)), LockTTL)
//				logger.Info.Printf("### SKIP and set offset to %d %s", offset, id)
//				return
//			}
//			logger.Info.Printf("!!! skip is ignore due to offset is %v, cluster size is %d", offset, sy.cluster.Size)
//		}
//
//		// Get bundle from WatcherPrefix
//		bundle, err := sy.etcdCli.GetSingleValue(WatcherPrefix + id)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		//fmt.Println("### Bundle is", bundle)
//		// Bundle goes to nil which means the key does not exist or error occurred
//		// We attempt to delete the credential key and left the lease for self-deleting
//		if bundle == nil {
//			sy.etcdCli.DeleteKey(CredPrefix + id)
//			logger.Info.Printf("Sync stoped due to missing key: %s%s", WatcherPrefix, id)
//			return
//		}
//		// Call sts AssumeRole
//		credential, err := sy.stsCli.AssumeRoleMock(bundle)
//		// Succeed
//		if err == nil {
//			// Set credential data to CredPrefix
//			err = sy.etcdCli.SetSingleValue(CredPrefix+id, string(credential)+fmt.Sprintf("%v", sy.cluster.Pid))
//			if err != nil {
//				logger.Error.Print(err)
//				return
//			}
//
//			// Set flag to 1 so that following n-1 (n = cluster size) watchers could escape
//			// The flag will be deleted after LockTTL
//			err = sy.etcdCli.SetSingleValueWithLease(FlagPrefix+id, strconv.Itoa(int(1)), LockTTL)
//			if err != nil {
//				logger.Error.Print(err)
//				return
//			}
//		} else {
//			logger.Error.Print(err)
//		}
//
//		//fmt.Println("### credential write to", CredPrefix+id)
//		// Set ttl to KeeperPrefix
//		err = sy.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, 10)
//		if err != nil {
//			logger.Error.Print(err)
//			return
//		}
//		logger.Info.Printf("@@@ SYNC successfully on id:%v %s", sy.cluster.Pid, id)
//	}()
//}
