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
	"cred/pkg/endpoint"
	"cred/pkg/processor"
	"cred/pkg/service"
	"cred/pkg/transport"
	"flag"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	loggerUtils "cred/utils/logger"
)

const (
	EnvPort    = "CRED_PORT"
	EnvMetaUrl = "CRED_META_URL"
	EnvSTSUrl  = "CRED_STS_URL"
	EnvTTL     = "CRED_SYNC_TTL"
)

type Config struct {
	Mock bool
}

var config Config

func init() {
	flag.BoolVar(&config.Mock, "mock", false, "enable mock mode")
}

func main() {
	flag.Parse()
	chans := processor.DefaultChans

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if config.Mock == false {
		runHttpServer(chans)
	}

	runProcessor(chans, config)

	for {
		select {
		case err := <-chans.ErrChan:
			loggerUtils.Error.Print(err)
		case s := <-signalChan:
			loggerUtils.Error.Printf("Captured %v. Exiting...", s)
			close(chans.DoneChan)
		case <-chans.DoneChan:
			os.Exit(0)
		}
	}

}

func runHttpServer(chans *processor.Chans) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		service     = service.NewCred(chans.SyncChan)
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

		loggerUtils.Info.Printf("Starting HTTP server at port %s", port)
		err := http.ListenAndServe(port, httpHandler)
		if err != nil {
			loggerUtils.Error.Print(err)
			close(chans.DoneChan)
		}
	}()
}

func runProcessor(chans *processor.Chans, config Config) {
	metaUrl := strings.Split(os.Getenv("CRED_META_URL"), ",")
	etcdCli, err := etcdv3.NewEtcdClient(metaUrl)

	if err != nil {
		loggerUtils.Error.Print(err)
		close(chans.DoneChan)
	}

	stsUrl := os.Getenv(EnvSTSUrl)
	stsCli := sts.NewStsClient(stsUrl)

	ttl, err := strconv.ParseInt(os.Getenv(EnvTTL), 10, 64)
	if err != nil {
		ttl = int64(20)
	}

	cluster := &processor.Cluster{Mock: config.Mock}

	reg := processor.NewRegister(etcdCli, cluster)
	reg.Process()

	sync := processor.NewSync(etcdCli, stsCli, ttl, cluster)
	sync.Process()

	watcher := processor.NewWatcher(etcdCli)
	watcher.Process()

	keeper := processor.NewKeeper(etcdCli)
	keeper.Process()

	loggerUtils.Info.Printf("Cluster id is: %d", cluster.Pid)
	loggerUtils.Info.Printf("Metadata url is: %s", metaUrl)
	loggerUtils.Info.Printf("STS url is: %s", stsUrl)
	loggerUtils.Info.Printf("Sync ttl is: %d", ttl)
}
