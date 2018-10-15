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
	"cred/utils/logger"
	"strings"
	"time"
)

// Keeper monitoring the /iam/lease/
// To keep tracking the lifecycle/timeout of temporary credential
type keeper struct {
	cli *etcdv3.Client
	wp  *workerPool
}

func NewKeeper(cli *etcdv3.Client, wp *workerPool) Processor {
	return &keeper{cli, wp}
}

func (ke *keeper) Process() {
	go ke.cli.WatchPrefix(KeeperPrefix, func(t int32, k, v []byte) error {
		switch t {
		case 0:
			//fmt.Println("Put event:", string(k), string(v))
		case 1:
			//fmt.Println("Delete event:", string(k))
			if !ke.wp.Serve(strings.TrimPrefix(string(k), KeeperPrefix)) {
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
		return nil
	})
}
