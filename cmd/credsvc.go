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

package main

import (
	"cred/pkg/backend/etcdv3"
	"cred/pkg/backend/sts"
	"cred/pkg/processor"
	"fmt"
	"github.com/go-kit/kit/log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	EnvPort    = "CRED_PORT"
	EnvMetaUrl = "CRED_META_URL"
	EnvSTSUrl  = "CRED_STS_URL"
	EnvTTL     = "CRED_SYNC_TTL"
)

func main() {
	chans := processor.DefaultChans

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	//var (
	//	service     = service.NewCred(chans.SyncChan)
	//	endpoints   = endpoint.MakeCredEndpoints(service, logger)
	//	httpHandler = transport.NewHttpHandler(endpoints, logger)
	//)

	go func() {
		port := os.Getenv(EnvPort)

		if port == "" {
			port = ":9011"
		} else {
			if !strings.HasPrefix(port, ":") {
				port = ":" + port
			}
		}

		//fmt.Println("Starting HTTP server at port", port)
		//err := http.ListenAndServe(port, httpHandler)
		//if err != nil {
		//	fmt.Println(err)
		//	close(chans.DoneChan)
		//}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	metasvc := []string{os.Getenv(EnvMetaUrl)}
	etcdCli, err := etcdv3.NewEtcdClient(metasvc)

	if err != nil {
		fmt.Println(err)
		close(chans.DoneChan)
	}

	stssvc := os.Getenv(EnvSTSUrl)
	stsCli := sts.NewStsClient(stssvc)

	ttl, err := strconv.ParseInt(os.Getenv(EnvTTL), 10, 64)
	if err != nil {
		ttl = int64(20)
	}

	cluster := &processor.Cluster{}

	reg := processor.NewRegister(etcdCli, cluster)
	reg.Process()

	sync := processor.NewSync(etcdCli, stsCli, ttl, cluster)
	sync.Process()

	watcher := processor.NewWatcher(etcdCli)
	watcher.Process()

	keeper := processor.NewKeeper(etcdCli)
	keeper.Process()

	fmt.Println("Cluster id is:", cluster.Pid)
	fmt.Println("Metadata url is:", metasvc)
	fmt.Println("STS url is:", stssvc)
	fmt.Println("Sync ttl is:", ttl)

	for {
		select {
		case err := <-chans.ErrChan:
			fmt.Println(err.Error())
		case s := <-signalChan:
			fmt.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			close(chans.DoneChan)
		case <-chans.DoneChan:
			os.Exit(0)
		}
	}

}
