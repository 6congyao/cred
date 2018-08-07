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
	"cred/pkg/endpoint"
	"cred/pkg/processor"
	"cred/pkg/service"
	"cred/pkg/transport"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	EnvPort    = "CRED_PORT"
	EnvMetaUrl = "CRED_META_URL"
)

func main() {
	ec := make(chan error)
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		service     = service.NewCred()
		endpoints   = endpoint.MakeCredEndpoints(service, logger)
		httpHandler = transport.NewHttpHandler(endpoints, logger)
	)

	go func() {
		port := os.Getenv(EnvPort)

		if port == "" {
			port = ":9011"
		} else {
			if !strings.HasPrefix(port, ":") {
				port = ":" + port
			}
		}

		fmt.Println("Starting HTTP server at port", port)
		ec <- http.ListenAndServe(port, httpHandler)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	//machines := []string{"http://139.198.177.151:2379"}
	//machines := []string{"http://139.198.120.106:2379"}

	//time.Sleep(5* time.Second)
	//cli.SetValues()
	//cli.GetValues()

	machines := []string{os.Getenv(EnvMetaUrl)}
	client, err := etcdv3.NewEtcdClient(machines)

	if err != nil {
		ec <- err
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)

	watcher := processor.NewWatcher(client, stopChan, doneChan)
	watcher.Process()

	keeper := processor.NewKeeper(client, stopChan, doneChan)
	keeper.Process()

	select {
	case err := <-ec:
		fmt.Println(err.Error())
	case s := <-signalChan:
		fmt.Println(fmt.Sprintf("Captured %v. Exiting...", s))
	}
}
