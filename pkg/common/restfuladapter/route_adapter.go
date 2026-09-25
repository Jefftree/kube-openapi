package restfuladapter

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/kube-openapi/pkg/common"
)

var _ common.Route = &RouteAdapter{}

// RouteAdapter adapts a restful.Route to common.Route.
type RouteAdapter struct {
	Route *restful.Route
}

func (r *RouteAdapter) StatusCodeResponses() []common.StatusCodeResponse {
	// go-restful uses the ResponseErrors field to contain both error and regular responses.
	if len(r.Route.ResponseErrors) == 0 {
		return nil
	}
	responses := make([]common.StatusCodeResponse, len(r.Route.ResponseErrors))
	errVals := make([]restful.ResponseError, len(r.Route.ResponseErrors))
	adapters := make([]ResponseErrorAdapter, len(r.Route.ResponseErrors))
	i := 0
	for _, res := range r.Route.ResponseErrors {
		errVals[i] = res
		adapters[i] = ResponseErrorAdapter{&errVals[i]}
		responses[i] = &adapters[i]
		i++
	}
	return responses
}

func (r *RouteAdapter) OperationName() string {
	return r.Route.Operation
}

func (r *RouteAdapter) Method() string {
	return r.Route.Method
}

func (r *RouteAdapter) Path() string {
	return r.Route.Path
}

func (r *RouteAdapter) Parameters() []common.Parameter {
	if len(r.Route.ParameterDocs) == 0 {
		return nil
	}
	params := make([]common.Parameter, len(r.Route.ParameterDocs))
	adapters := make([]ParamAdapter, len(r.Route.ParameterDocs))
	for i, rParam := range r.Route.ParameterDocs {
		adapters[i].Param = rParam
		params[i] = &adapters[i]
	}
	return params
}

func (r *RouteAdapter) Description() string {
	return r.Route.Doc
}

func (r *RouteAdapter) Consumes() []string {
	return r.Route.Consumes
}

func (r *RouteAdapter) Produces() []string {
	return r.Route.Produces
}

func (r *RouteAdapter) Metadata() map[string]interface{} {
	return r.Route.Metadata
}

func (r *RouteAdapter) RequestPayloadSample() interface{} {
	return r.Route.ReadSample
}

func (r *RouteAdapter) ResponsePayloadSample() interface{} {
	return r.Route.WriteSample
}
