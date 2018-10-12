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
	"fmt"
	"os"
	"time"
)

// Register attempt to maintain the node info in cluster
type register struct {
	cli      *etcdv3.Client
	cluster  *Cluster
	doneChan chan struct{}
}

func NewRegister(cli *etcdv3.Client, cluster *Cluster) Processor {
	return &register{cli, cluster, DefaultChans.DoneChan}
}

func (re *register) Process() {
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
		logger.Info.Printf("Cluster size changes to: %d", re.cluster.Size)
		return nil
	})
	logger.Info.Printf("Cluster Regkey is: %s", re.cluster.RegKey)
	logger.Info.Printf("Cluster size is: %d", re.cluster.Size)
}

func (re *register) doRegister() {
	key := fmt.Sprintf("%s%x", RegPrefix, re.cluster.Pid)
	for {
		_, regkey, err := re.cli.Register(key, LockTTL)
		if err != nil {
			logger.Error.Print(err)
			time.Sleep(time.Duration(1) * time.Second)
		} else {
			re.cluster.RegKey = regkey
			break
		}
	}
}
