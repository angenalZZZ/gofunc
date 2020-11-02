package http

import "github.com/go-resty/resty/v2"

var (
	NewRestClient  = resty.New
	NewRestRequest = NewRestClient().R
)

func NewRestJsonRequest(token ...string) (r *resty.Request) {
	r = NewRestRequest()
	// POST Struct, default is JSON content type. No need to set one
	r.SetHeader("Content-Type", "application/json")
	r.SetHeader("Accept", "application/json")
	if len(token) > 0 && token[0] != "" {
		r.SetAuthToken(token[0])
	}
	return
}

func NewRestFormRequest(token ...string) (r *resty.Request) {
	r = NewRestRequest()
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	if len(token) > 0 && token[0] != "" {
		r.SetAuthToken(token[0])
	}
	return
}
