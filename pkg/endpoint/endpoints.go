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

package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"cred/pkg/service"
)

// Endpoints collects all of the endpoints that compose a sts service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	AssumeRoleEndpoint endpoint.Endpoint
}

func MakeStsEndpoints(svc service.Service, logger log.Logger) Endpoints {
	assumeroleEdp := makeAssumeRoleEndpoint(svc)
	assumeroleEdp = LoggingMiddleware(log.With(logger, "method", "AssumeRole"))(assumeroleEdp)
	// todo: Prometheus
	//assumeroleEdp = InstrumentingMiddleware(duration.With("method", "AssumeRole"))(assumeroleEdp)
	return Endpoints{
		AssumeRoleEndpoint: assumeroleEdp,
	}
}

func makeAssumeRoleEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(AssumeRoleRequest)
		token, err := svc.AssumeRole(ctx, req.RoleQrn, req.ExternalId)
		if err != nil {
			return nil, err
		}
		return AssumeRoleResponse{
			Token: token,
			Err:   err,
		}, nil
	}
}

// Failer is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so if they've
// failed, and if so encode them using a separate write path based on the error.
type Failer interface {
	Failed() error
}

type AssumeRoleRequest struct {
	DurationSeconds int64  `json:"duration_seconds"`
	ExternalId      string `json:"external_id"`
	Policy          string `json:"policy"`
	RoleQrn         string `json:"role_qrn"`
}

type AssumeRoleResponse struct {
	Token string `json:"token"`
	Err   error  `json:"err"`
}

func (r AssumeRoleResponse) Failed() error { return r.Err }
