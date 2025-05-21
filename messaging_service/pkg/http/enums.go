

package httpclient

type HTTPMethod string

const (
	MethodGet    HTTPMethod = "GET"
	MethodPost   HTTPMethod = "POST"
	MethodPut    HTTPMethod = "PUT"
	MethodPatch  HTTPMethod = "PATCH"
	MethodDelete HTTPMethod = "DELETE"
)

type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeXML  ContentType = "application/xml"
)
