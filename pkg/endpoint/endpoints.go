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
	"cred/pkg/service"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Endpoints collects all of the endpoints that compose a Cred service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	HealthEndpoint     endpoint.Endpoint
	RefreshCredentialEndpoint endpoint.Endpoint
}

func MakeCredEndpoints(svc service.Service, logger log.Logger) Endpoints {
	healthEdp := makeHealthEndpoint(svc)
	healthEdp = LoggingMiddleware(log.With(logger, "method", "health"))(healthEdp)
	refreshCredentialEdp := makeRefreshCredentialEndpoint(svc)
	refreshCredentialEdp = LoggingMiddleware(log.With(logger, "method", "RefreshCredential"))(refreshCredentialEdp)
	// todo: Prometheus
	//refreshCredentialEdp = InstrumentingMiddleware(duration.With("method", "RefreshCredential"))(refreshCredentialEdp)
	return Endpoints{
		HealthEndpoint:     healthEdp,
		RefreshCredentialEndpoint: refreshCredentialEdp,
	}
}

func makeHealthEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		err = svc.Health(ctx)
		return HealthResponse{}, err
	}
}

func makeRefreshCredentialEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RefreshCredentialRequest)
		err = svc.RefreshCredential(ctx, req.Target)
		if err != nil {
			return nil, err
		}
		return RefreshCredentialResponse{
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

type HealthRequest struct{}

type HealthResponse struct{}

type RefreshCredentialRequest struct {
	Target interface{}  `json:"target"`
}

type RefreshCredentialResponse struct {
	Err   error  `json:"err"`
}

func (r RefreshCredentialResponse) Failed() error { return r.Err }
