package olives

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Request struct {
	router  *gin.Engine
	method  string
	headers http.Header
	path    map[string]string
	query   url.Values
	body    io.Reader
	cookies map[string][]string
	err     error
}

func NewRequest() *Request {
	// Disable logging output to keep test logs clear
	gin.SetMode(gin.ReleaseMode)

	return &Request{
		method:  "GET",
		router:  gin.New(),
		headers: make(http.Header),
		path:    make(map[string]string),
		query:   make(url.Values),
		cookies: make(map[string][]string),
	}
}

func (r *Request) WithMethod(method string) *Request {
	r.method = method
	return r
}

func (r *Request) WithHeader(key, value string) *Request {
	r.headers.Add(key, value)
	return r
}

func (r *Request) WithQuery(key, value string) *Request {
	r.query.Add(key, value)
	return r
}

func (r *Request) WithPath(key, value string) *Request {
	r.path[key] = value
	return r
}

func (r *Request) WithBody(rdr io.Reader) *Request {
	r.body = rdr
	return r
}

func (r *Request) WithBytes(body []byte) *Request {
	r.body = bytes.NewReader(body)
	return r
}

func (r *Request) WithJSON(body interface{}) *Request {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		r.err = err
		return r
	}

	return r.WithBytes(bodyBytes)
}

func (r *Request) WithCookie(name, contents string) *Request {
	r.cookies[name] = append(r.cookies[name], contents)
	return r
}

func (r *Request) Send(handlers ...gin.HandlerFunc) (*httptest.ResponseRecorder, error) {
	handlerPath, requestPath := r.paths()
	r.router.Handle(r.method, handlerPath, handlers...)

	url := fmt.Sprintf("%s?%s", requestPath, r.query.Encode())
	req, err := http.NewRequest(r.method, url, r.body)
	if err != nil {
		return nil, err
	}

	req.Header = r.headers
	for name, vals := range r.cookies {
		for _, val := range vals {
			req.AddCookie(&http.Cookie{Name: name, Value: val})
		}
	}

	w := httptest.NewRecorder()
	r.router.ServeHTTP(w, req)
	return w, nil
}

func (r *Request) paths() (handlerPath, requestPath string) {
	var hp, rp strings.Builder
	fmt.Fprintf(&hp, "/")
	fmt.Fprintf(&rp, "/")

	for param, val := range r.path {
		fmt.Fprintf(&hp, ":%s/", param)
		fmt.Fprintf(&rp, "%s/", val)
	}

	handlerPath = hp.String()
	requestPath = rp.String()
	return
}
