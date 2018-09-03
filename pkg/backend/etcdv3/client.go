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
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"time"
	"cred/utils/logger"
)

type EventHandler func(t int32, k, v []byte) error

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

func (c *Client) WatchPrefix(prefix string, eh EventHandler) error {
	rch := c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	logger.Info.Printf("Watch created on %s", prefix)
	for {
		for wresp := range rch {
			for _, event := range wresp.Events {
				eh(int32(event.Type), event.Kv.Key, event.Kv.Value)
				//switch event.Type {
				//case mvccpb.PUT:
				//	fmt.Println("Put event :", string(event.Kv.Key), string(event.Kv.Value))
				//	obj := make(map[string]interface{})
				//	err := json.Unmarshal(event.Kv.Value, &obj)
				//	if err != nil {
				//		fmt.Println("error :", err.Error())
				//	}
				//	fmt.Println("Put obj :", obj)
				//case mvccpb.DELETE:
				//	fmt.Println("Delete event:", string(event.Kv.Key))
				//}
			}
		}

		// Reconnect while lost or closed
		logger.Warn.Printf("Connection lost on %s", prefix)
		time.Sleep(time.Duration(1) * time.Second)
		rch = c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	}

	return nil
}

func (c *Client) GetSingleValue(key string) ([]byte, error) {
	var ret []byte = nil
	resp, err := c.client.Get(context.Background(), key)

	if err != nil {
		logger.Error.Print(err)
		return nil, err
	}

	for _, v := range resp.Kvs {
		ret = v.Value
	}
	return ret, nil
}

func (c *Client) GetKeyNumber(prefix string) (int64, error) {
	resp, err := c.client.Get(context.Background(), prefix, clientv3.WithPrefix())

	if err != nil {
		logger.Error.Print(err)
		return 0, err
	}

	return resp.Count, nil
}

func (c *Client) SetSingleValue(key, value string) error {
	_, err := c.client.Put(context.Background(), key, value)

	if err != nil {
		logger.Error.Print(err)
		return err
	}
	return nil
}

func (c *Client) SetSingleValueWithLease(key, value string, ttl int64) error {
	resp, err := c.client.Grant(context.Background(), ttl)
	if err != nil {
		logger.Error.Print(err)
		return err
	}

	_, err = c.client.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))

	if err != nil {
		logger.Error.Print(err)
		return err
	}
	return nil
}

func (c *Client) DeleteKey(key string) error {
	resp, err := c.client.Delete(context.Background(), key)

	if err != nil {
		logger.Error.Print(err)
		return err
	}
	for _, v := range resp.PrevKvs {
		logger.Info.Print(v.Key, v.Value)
	}
	return nil
}

func (c *Client) Lock(key string, ttl int64) (*concurrency.Session, *concurrency.Mutex, error) {
	resp, err := c.client.Grant(context.Background(), ttl)
	if err != nil {
		logger.Error.Print(err)
		return nil, nil, err
	}

	s, err := concurrency.NewSession(c.client, concurrency.WithLease(resp.ID))
	if err != nil {
		logger.Error.Print(err)
		return nil, nil, err
	}

	m := concurrency.NewMutex(s, key)
	if err := m.Lock(context.Background()); err != nil {
		logger.Error.Print(err)
		return nil, nil, err
	}
	return s, m, nil
}

func (c *Client) Unlock(s *concurrency.Session, m *concurrency.Mutex) error {
	defer s.Close()
	if err := m.Unlock(context.Background()); err != nil {
		logger.Error.Print(err)
		return err
	}
	return nil
}

func (c *Client) Register(key string, ttl int64) (*concurrency.Session, string, error) {
	resp, err := c.client.Grant(context.Background(), ttl)
	if err != nil {
		logger.Error.Print(err)
		return nil, "", err
	}

	s, err := concurrency.NewSession(c.client, concurrency.WithLease(resp.ID))
	if err != nil {
		logger.Error.Print(err)
		return nil, "", err
	}

	myKey := fmt.Sprintf("%s%x", key+"/", s.Lease())

	_, err = c.client.Put(context.Background(), myKey, "", clientv3.WithLease(s.Lease()))
	if err != nil {
		logger.Error.Print(err)
		return nil, "", err
	}

	return s, myKey, nil
}
