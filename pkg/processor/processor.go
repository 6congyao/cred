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
	"strings"
)

const (
	WatcherPrefix = "/iam/instance-profile/"
	KeeperPrefix  = "/iam/lease/"
	CredPrefix    = "/iam/credential/"

	MaxQueueSize = 20
)

type Processor interface {
	Process()
}

// Watcher monitoring the /iam/instance-profile/
type watcher struct {
	cli      *etcdv3.Client
	syncChan chan string
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
type keeper struct {
	cli      *etcdv3.Client
	syncChan chan string
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

// Sync attempt to write the data to /iam/credential/
type sync struct {
	etcdCli  *etcdv3.Client
	stsCli   *sts.Client
	syncChan chan string
	ttl      int64
}

func NewSync(etcdCli *etcdv3.Client, stsCli *sts.Client, ttl int64) Processor {
	return &sync{etcdCli, stsCli, DefaultChans.SyncChan, ttl}
}

func (sy sync) Process() {
	go func() {
		fmt.Println("Waiting on sync...")
		for v := range sy.syncChan {
			fmt.Println("SyncChan got update on:", v)
			sy.doSync(v)
		}
	}()
}

func (sy sync) doSync(id string) {
	go func() {
		rid := "#" + fmt.Sprint(rand.Int63()%32768)
		// Get bundle from WatcherPrefix
		bundle, err := sy.etcdCli.GetSingleValue(WatcherPrefix + id)
		if err != nil {
			fmt.Println(err)
		}
		// Bundle goes to nil which means the key does not exist or error occurred
		// We attempt to delete the credential key and left the lease for self-deleting
		if bundle == nil {
			sy.etcdCli.DeleteKey(CredPrefix + id)
			fmt.Println("Sync stoped due to missing key:", WatcherPrefix+id)
			return
		}
		// Call sts AssumeRole
		credential, err := sy.stsCli.AssumeRole(bundle)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Set data to CredPrefix
		err = sy.etcdCli.SetSingleValue(CredPrefix+id, string(credential))
		if err != nil {
			fmt.Println(err)
			return
		}
		// Set ttl to KeeperPrefix
		err = sy.etcdCli.SetSingleValueWithLease(KeeperPrefix+id, rid, sy.ttl)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Sync successfully on random id:", rid)
	}()
}

type chans struct {
	DoneChan chan struct{}
	ErrChan  chan error
	SyncChan chan string
}

func NewChans() *chans {
	return &chans{make(chan struct{}), make(chan error, MaxQueueSize), make(chan string, MaxQueueSize)}
}

var DefaultChans = NewChans()
