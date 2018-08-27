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

package sts

import (
	"bytes"
	"cred/utils/pester"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"cred/utils/logger"
)

type Client struct {
	client  *pester.Client
	baseUrl string
}

func NewStsClient(baseUrl string) *Client {
	pst := pester.New()
	pst.MaxRetries = 3
	pst.KeepLog = true

	return &Client{pst, baseUrl}
}

func (c Client) AssumeRole(payload []byte) ([]byte, error) {
	resp, err := c.client.Post(c.baseUrl+"/role", "application/json", bytes.NewReader(payload))

	if err != nil {
		//fmt.Println(time.Now().Format("2006-01-02 15:04:05"),"assume role error:", err)
		logger.Error.Printf("assume role error:%s", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("AssumeRole error with status [%d]", resp.StatusCode))
	}
	credential, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return credential, nil
}

func (c Client) AssumeRoleMock(payload []byte) ([]byte, error) {
	time.Sleep(time.Duration(5) * time.Second)
	return payload, nil
}
