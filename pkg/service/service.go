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
)

type Service interface {
	AssumeRole(ctx context.Context, role, principal string) (string, error)
	Healthy(ctx context.Context) error
}

type cred struct{}

func NewCred() Service {
	return &cred{}
}

func (c cred) AssumeRole(ctx context.Context, role, principal string) (string, error) {
	// Firstly we check the resource-based-policy for the role if it could be assumed
	//err := client.Evaluate(ctx, role, principal)
	//if err != nil {
	//	return "", err
	//}
	//
	////Attempt to get an id token from issuer
	//err = client.Issue(ctx)
	//if err != nil {
	//	return "", err
	//}
	//
	//fmt.Println("evaluation was allowed")

	return "", nil
}

func (c cred) Healthy(ctx context.Context) error {
	return nil
}
