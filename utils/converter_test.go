/*******************************************************************************
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
 ******************************************************************************/

package utils

import (
	"testing"
	"log"
)

const checkPass = "\u2713"
const checkFail = "\u2717"

func TestItoS(t *testing.T) {
	log.Println("Test ToStringSlice")
	for k, c := range []struct {
		input  interface{}
		output []string
	}{

		{[]interface{}{0, "bbb"}, []string{"0", "bbb"}},
		{[]interface{}{false, "bbb"}, []string{"false" ,"bbb"}},
		{[]interface{}{true, "bbb"}, []string{"true" ,"bbb"}},
		{[]interface{}{66666, "bbb"}, []string{"66666", "bbb"}},
		{"ccccc", []string{"ccccc"}},
		{[]interface{}{"qrn:user/max", "qrn:group/dev"}, []string{"qrn:user/max", "qrn:group/dev"}},

		{[]interface{}{"qrn:qws:s3:::max/*",
			"qrn:qws:s3:::min/*"}, []string{"qrn:qws:s3:::max/*",
			"qrn:qws:s3:::min/*"}},

		{[]interface{}{"qrn:qws:s3:::max/*",
			"qrn:qws:s3:::min/*"}, []string{"qrn:qws:s3:::max/*",
			"qrn:qws:s3:::min/*"}},

		{[]interface{}{"qrn:qws:s3:::max/*,qrn:qws:s3:::min/*"}, []string{"qrn:qws:s3:::max/*,qrn:qws:s3:::min/*"}},
		{"qrn:qws:s3:::max/*,qrn:qws:s3:::min/*", []string{"qrn:qws:s3:::max/*,qrn:qws:s3:::min/*"}},
		{[]interface{}{2147483646, "-2147483646"}, []string{"2147483646", "-2147483646"}},
		{[]interface{}{2147483648, -2147483647}, []string{"2147483648", "-2147483647"}}, //pow(2,31)
		{[]interface{}{4294967296, "-4294967296"}, []string{"4294967296", "-4294967296"}},//pow(2,32)
		{[]interface{}{4294967297, "-4294967297"}, []string{"4294967297", "-4294967297"}},//pow(2,32)+1
		{[]interface{}{9294967297, "-4294967297"}, []string{"9294967297", "-4294967297"}},
		{[]interface{}{1.1, "-0.0"}, []string{"1.1", "-0.0"}},
		{[]interface{}{float64(1000000), "-0.0"}, []string{"1000000", "-0.0"}},
		{[]interface{}{1000000.1, "-0.0"}, []string{"1000000.1", "-0.0"}},//fixed


		{"s3:GetObject", []string{"s3:GetObject"}},

	} {
		if !compare(ToStringSlice(c.input), c.output) {
			t.Errorf("Error! case %d input:%#v output:%#v correct:%s %s",k+1,c.input, ToStringSlice(c.input), c.output, checkFail)
		}else {
			//t.Logf("PASS! case %d input:%#v output:%#v correct:%s %s",k+1,c.input, ToStringSlice(c.input), c.output, checkPass)
		}
	}

}
func compare(s, t []string) bool {
	if s == nil && t == nil {
		return true
	}
	if len(s) != len(t) {
		return false
	}
	for k, _ := range s {
		if s[k] != t[k] {
			return false
		}
	}
	return true
}
