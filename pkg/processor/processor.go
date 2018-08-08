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
	"strings"
)

const (
	WatcherKey = "/iam/instance-profile/"
	KeeperKey  = "/iam/credential/"
)

type Processor interface {
	Process()
}

// Watcher monitoring the /iam/instance-profile
type watcher struct {
	cli      *etcdv3.Client
	credChan chan string
}

func NewWatcher(cli *etcdv3.Client, credChan chan string) Processor {
	return &watcher{cli, credChan}
}

func (wa watcher) Process() {
	go wa.cli.WatchPrefix(WatcherKey, func(t int32, k, v []byte) error {
		switch t {
		case 0:
			fmt.Println("Put event:", string(k), string(v))
			wa.credChan <- strings.TrimPrefix(string(k), WatcherKey)
		case 1:
			fmt.Println("Delete event:", string(k))
		}
		return nil
	})
}

// Keeper monitoring the /iam/credential
type keeper struct {
	cli      *etcdv3.Client
	credChan chan string
}

func NewKeeper(cli *etcdv3.Client, credChan chan string) Processor {
	return &keeper{cli, credChan}
}

func (ke keeper) Process() {
	go ke.cli.WatchPrefix(KeeperKey, func(t int32, k, v []byte) error {
		switch t {
		case 0:
			fmt.Println("Put event:", string(k), string(v))
		case 1:
			fmt.Println("Delete event:", string(k))
			ke.credChan <- strings.TrimPrefix(string(k), KeeperKey)
		}
		return nil
	})
}

// Sync getting and putting the data
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
			fmt.Println("credchan:", v)
			sy.doSync(v)
		}
	}()
}

func (sy sync) doSync(id string) {
	go func() {
		//fmt.Println("start sync")
		//defer fmt.Println("stop sync")
		//time.Sleep(time.Duration(10)*time.Second)

		bundle, err := sy.cli.GetSingleValue(WatcherKey + id)
		if err != nil {
			fmt.Println(err)
		}
		err = sy.cli.SetSingleValue(KeeperKey+id, string(bundle))
		if err != nil {
			fmt.Println(err)
		}
	}()
}
