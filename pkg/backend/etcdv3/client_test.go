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
	"os"
	"testing"
	"time"
)

const (
	ttl = 3
)

var client *Client

func init() {
	machines := []string{os.Getenv("CRED_META_URL")}
	client, _ = NewEtcdClient(machines)
}

func TestSetSingleValueWithLease(t *testing.T) {
	key := "/iam/credential/testttl"

	expected := "expected"
	client.SetSingleValueWithLease(key, expected, ttl)
	actual, _ := client.GetSingleValue(key)
	if string(actual) == expected {
		time.Sleep(time.Duration(ttl+2) * time.Second)
		actual, _ = client.GetSingleValue(key)
		if actual != nil {
			t.Errorf("Expected the result to be nil but instead got %s", actual)
		}
	}
}

func TestSetSingleValue(t *testing.T) {
	key := "/iam/credential/testset"

	expected := "expected"
	client.SetSingleValue(key, expected)
	actual, _ := client.GetSingleValue(key)
	if string(actual) != expected {
		t.Errorf("Expected the result to be %s but instead got %s", expected, actual)
	}
}

func TestGetWithNoExistKey(t *testing.T) {
	key := "/noexistkey"

	var expected []byte = nil
	actual, _ := client.GetSingleValue(key)
	if string(actual) != string(expected) {
		t.Errorf("Expected the result to be %s but instead got %s", expected, actual)
	}
}

func TestDelete(t *testing.T) {
	//key1 := "/notexistkey"
	key2 := "/iam/instance-profile/i-fff/area/pro/1/hunan"
	var expected error = nil
	actual := client.DeleteKey(key2)
	if actual != expected {
		t.Errorf("Expected the result to be %v but instead got %v", expected, actual)
	}
}
