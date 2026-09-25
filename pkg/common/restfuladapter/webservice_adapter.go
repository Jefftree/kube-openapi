package restfuladapter

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/kube-openapi/pkg/common"
)

var _ common.RouteContainer = &WebServiceAdapter{}

// WebServiceAdapter adapts a restful.WebService to common.RouteContainer.
type WebServiceAdapter struct {
	WebService *restful.WebService
}

func (r *WebServiceAdapter) RootPath() string {
	return r.WebService.RootPath()
}

func (r *WebServiceAdapter) PathParameters() []common.Parameter {
	wsParams := r.WebService.PathParameters()
	if len(wsParams) == 0 {
		return nil
	}
	params := make([]common.Parameter, len(wsParams))
	adapters := make([]ParamAdapter, len(wsParams))
	for i, rParam := range wsParams {
		adapters[i].Param = rParam
		params[i] = &adapters[i]
	}
	return params
}

func (r *WebServiceAdapter) Routes() []common.Route {
	wsRoutes := r.WebService.Routes()
	if len(wsRoutes) == 0 {
		return nil
	}
	routes := make([]common.Route, len(wsRoutes))
	adapters := make([]RouteAdapter, len(wsRoutes))
	for i := range wsRoutes {
		adapters[i].Route = &wsRoutes[i]
		routes[i] = &adapters[i]
	}
	return routes
}
