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

package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

type Client struct {
	client *clientv3.Client
}

func NewEtcdClient(machines []string) (*Client, error) {
	cfg := clientv3.Config{
		Endpoints:            machines,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return &Client{}, err
	}
	return &Client{cli}, nil
}

func (c *Client) WatchPrefix(prefix string) error {
	rch := c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	fmt.Println("Watch created on", prefix)
	for {
		for wresp := range rch {
			for _, event := range wresp.Events {
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("Put event :", string(event.Kv.Key), string(event.Kv.Value))
					obj := make(map[string]interface{})
					err := json.Unmarshal(event.Kv.Value, &obj)
					if err != nil {
						fmt.Println("error :", err.Error())
					}
					fmt.Println("Put obj :", obj)
				case mvccpb.DELETE:
					fmt.Println("Delete event:", string(event.Kv.Key))
				}
			}
		}

		// Reconnect while lost
		fmt.Println("Warning, connection lost on", prefix)
		time.Sleep(time.Duration(1) * time.Second)
		rch = c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	}

	return nil
}

func (c *Client) GetValues() error {
	kresp, err := c.client.Get(context.Background(), "/iam/instance-profile/", clientv3.WithPrefix())

	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, v := range kresp.Kvs {
		fmt.Println("Value is", string(v.Key), string(v.Value))
	}
	return nil
}

func (c *Client) SetValues() error {
	_, err := c.client.Put(context.Background(), "/iam/instance-profile/i-678", "{'name': '123', 'city': 'wuhan'}")

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
