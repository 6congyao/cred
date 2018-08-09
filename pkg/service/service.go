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

package service

import (
	"context"
	"cred/utils"
)

type Service interface {
	RefreshCredential(ctx context.Context, target interface{}) error
	Health(ctx context.Context) error
}

type cred struct {
	credChan chan string
}

func NewCred(credChan chan string) Service {
	return &cred{credChan}
}

func (c cred) RefreshCredential(ctx context.Context, target interface{}) error {
	ids := utils.ToStringSlice(target)
	for _, v := range ids {
		c.credChan <- v
	}
	return nil
}

func (c cred) Health(ctx context.Context) error {
	return nil
}
