package app

import (
	"net/http"
)

const (
	apiPathRankerQuery = "/v1/ranker/query"
)

func (a *Application) getRouter() *http.ServeMux {

	// Create and configure routes
	serveMux := http.NewServeMux()

	// Status API
	{
		// serveMux.Handle(apiPathStatusGet, a.getCommonWrapperHandler(a.statusGetHandler()))
		serveMux.Handle(apiPathRankerQuery, a.getCommonWrapperHandler(a.rankerQueryHandler()))
	}

	return serveMux
}
