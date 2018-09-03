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
	"strconv"
	"testing"
	"strings"
)

const (
	ttl = 3
)

var client *Client
var client2 *Client
var c chan struct{}

func init() {
	machines := strings.Split(os.Getenv("CRED_META_URL"), ",")
	client, _ = NewEtcdClient(machines)
	client2, _ = NewEtcdClient(machines)
	c = make(chan struct{})
}

func TestGetKeyNumber(t *testing.T) {
	prefix := "/iam/credential/"
	//prefix := "/iam/instance-profile/"
	expected := int64(5)
	actual, _ := client.GetKeyNumber(prefix)
	if actual != expected {
		t.Errorf("Expected the result to be %v but instead got %v", expected, actual)
	}
}

//func TestLock(t *testing.T) {
//	fmt.Println("acquired lock 1")
//	prefix := "/iam/mutex/1"
//	s, m, err := client.Lock(prefix, 25)
//	if err != nil {
//		t.Errorf("Expected the lock successfully but got %s", err)
//	}
//	defer client.Unlock(s, m)
//	fmt.Println("Locked 1")
//
//	//time.Sleep(time.Duration(20)* time.Second)
//	//go func() {
//	//	defer close(c)
//	//	fmt.Println("acquired lock 2")
//	//	_, err := client.Lock(prefix, 20)
//	//	if err != nil{
//	//
//	//	}
//	//	fmt.Println("Locked 2")
//	//
//	//}()
////time.Sleep(time.Duration(10)* time.Second)
//	//client.Unlock(m1, 1)
//	fmt.Println("lock 1 released")
//	//err = client.Unlock(m, 1)
//	//if err != nil {
//	//	t.Errorf("Expected the lock successfully but got %s", err)
//	//}
//	//<-c
//}

//func TestLock2(t *testing.T) {
//	go func() {
//		defer close(c)
//		fmt.Println("acquired lock 2")
//		prefix := "/iam/mutex"
//		_, err := client2.Lock(prefix, 20)
//		if err != nil {
//			t.Errorf("Expected the lock successfully but got %s", err)
//		}
//		fmt.Println("Locked 2")
//		//err = client.Unlock(m, 1)
//		//if err != nil {
//		//	t.Errorf("Expected the lock successfully but got %s", err)
//		//}
//	}()
//	<-c
//}

func TestSetSingleValueWithLease(t *testing.T) {
	key := "/iam/flag/i-lu"

	expected := 3
	client.SetSingleValueWithLease(key, strconv.Itoa(expected), 30)
	value, _ := client.GetSingleValue(key)
	actual, _ := strconv.Atoi(string(value))
	if actual != expected {
		//time.Sleep(time.Duration(ttl+2) * time.Second)
		//actual, _ = client.GetSingleValue(key)
		//if actual != nil {
		//	t.Errorf("Expected the result to be nil but instead got %s", actual)
		//}
		t.Errorf("Expected the result to be %v but instead got %v", expected, actual)
	}
}

func TestSetSingleValue(t *testing.T) {
	key := "/iam/flag/i-lu"

	expected := 6
	client.SetSingleValue(key, strconv.Itoa(expected))
	value, _ := client.GetSingleValue(key)
	actual, _ := strconv.Atoi(string(value))
	if actual != expected {
		t.Errorf("Expected the result to be %s but instead got %s", expected, actual)
	}
}

func TestGetWithNoExistKey(t *testing.T) {
	key := "/iam/flag/i-lu"

	expected := 6
	actual, _ := client.GetSingleValue(key)
	v, _ := strconv.ParseInt(string(actual), 0, 0)

	i := int(v)
	//if string(actual) != string(expected) {
	//	t.Errorf("Expected the result to be %s but instead got %s", expected, actual)
	//}
	if i != expected {
		t.Errorf("Expected the result to be %v but instead got %v", expected, i)
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
