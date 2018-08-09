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
	"fmt"
	"math/rand"
	"strings"
)

const (
	WatcherPrefix = "/iam/instance-profile/"
	KeeperPrefix  = "/iam/lease/"
	CredPrefix    = "/iam/credential/"
)

type Processor interface {
	Process()
}

// Watcher monitoring the /iam/instance-profile/
type watcher struct {
	cli      *etcdv3.Client
	credChan chan string
}

func NewWatcher(cli *etcdv3.Client, credChan chan string) Processor {
	return &watcher{cli, credChan}
}

func (wa watcher) Process() {
	go wa.cli.WatchPrefix(WatcherPrefix, func(t int32, k, v []byte) error {
		// Trigger the sync immediately
		wa.credChan <- strings.TrimPrefix(string(k), WatcherPrefix)

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
	credChan chan string
}

func NewKeeper(cli *etcdv3.Client, credChan chan string) Processor {
	return &keeper{cli, credChan}
}

func (ke keeper) Process() {
	go ke.cli.WatchPrefix(KeeperPrefix, func(t int32, k, v []byte) error {
		switch t {
		case 0:
			//fmt.Println("Put event:", string(k), string(v))
		case 1:
			//fmt.Println("Delete event:", string(k))
			ke.credChan <- strings.TrimPrefix(string(k), KeeperPrefix)
		}
		return nil
	})
}

// Sync attempt to write the data to /iam/credential/
type sync struct {
	cli      *etcdv3.Client
	credChan chan string
}

func NewSync(cli *etcdv3.Client, credChan chan string) Processor {
	return &sync{cli, credChan}
}

func (sy sync) Process() {
	go func() {
		fmt.Println("Waiting on sync...")
		for v := range sy.credChan {
			fmt.Println("credchan got update on:", v)
			sy.doSync(v)
		}
	}()
}

func (sy sync) doSync(id string) {
	go func() {
		rid := "#" + fmt.Sprint(rand.Int63()%32768)
		// Get bundle from WatcherPrefix
		bundle, err := sy.cli.GetSingleValue(WatcherPrefix + id)
		if err != nil {
			fmt.Println(err)
		}
		// Bundle goes to nil which means the key does not exist or error occurred
		// We attempt to delete the credential key and left the lease for self-deleting
		if bundle == nil {
			sy.cli.DeleteKey(CredPrefix + id)
			fmt.Println("Sync stoped due to missing key:", WatcherPrefix+id)
			return
		}
		// Set data to CredPrefix
		err = sy.cli.SetSingleValue(CredPrefix+id, string(bundle)+rid)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Set ttl to KeeperPrefix
		err = sy.cli.SetSingleValueWithLease(KeeperPrefix+id, rid, 10)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Sync successfully on random id:", rid)
	}()
}
