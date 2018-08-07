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
	stopChan chan bool
	doneChan chan bool
}

func NewWatcher(cli *etcdv3.Client, stopChan, doneChan chan bool) Processor {
	return &watcher{cli, stopChan, doneChan}
}

func (w watcher) Process() {
	go w.cli.WatchPrefix(WatcherKey)
}

// Keeper monitoring the /iam/credential
type keeper struct {
	cli      *etcdv3.Client
	stopChan chan bool
	doneChan chan bool
}

func NewKeeper(cli *etcdv3.Client, stopChan, doneChan chan bool) Processor {
	return &keeper{cli, stopChan, doneChan}
}
func (k keeper) Process() {
	go k.cli.WatchPrefix(KeeperKey)
}
