package olives_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/itzamna314/olives"
)

func TestMethod(t *testing.T) {
	handler := func(verb *string) func(c *gin.Context) {
		return func(c *gin.Context) {
			*verb = c.Request.Method

			c.Status(http.StatusOK)
		}
	}

	testCases := []string{"GET", "POST", "PATCH", "PUT", "DELETE", "SILLY"}

	for _, method := range testCases {
		t.Run(method, func(t *testing.T) {
			var actualMethod string
			w, err := olives.NewRequest().
				WithMethod(method).
				Send(handler(&actualMethod))

			if err != nil {
				t.Errorf("Got unexpected error %s sending request method %s", err, method)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Got unexpected response status %d", w.Code)
			}

			if method != actualMethod {
				t.Errorf("Expected http request with method %s, but was %s", method, actualMethod)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	handler := func(hdr http.Header) func(c *gin.Context) {
		return func(c *gin.Context) {
			for key, vals := range c.Request.Header {
				hdr[key] = vals
			}

			c.Status(http.StatusOK)
		}
	}

	testCases := []struct {
		testName string
		header   http.Header
	}{
		{"empty", make(http.Header)},
		{"one", http.Header{"Content-Type": []string{"text/json"}}},
		{"two", http.Header{"X-Treats": []string{"gelato", "cake"}}},
		{"multi", http.Header{"X-Treats": []string{"gelato", "cake"}, "X-Names": []string{"diesel", "widow"}}},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			actualHeaders := make(http.Header)
			req := olives.NewRequest()
			for name, vals := range tt.header {
				for _, val := range vals {
					req = req.WithHeader(name, val)
				}
			}

			w, err := req.Send(handler(actualHeaders))
			if err != nil {
				t.Errorf("Unexpected olives error: %s", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Got unexpected response status: %d", w.Code)
			}

			mapSlicesEq(t, tt.header, actualHeaders)
		})
	}
}

func TestQuery(t *testing.T) {
	handler := func(qs url.Values) func(c *gin.Context) {
		return func(c *gin.Context) {
			for key, vals := range c.Request.URL.Query() {
				qs[key] = vals
			}

			c.Status(http.StatusOK)
		}
	}

	testCases := []struct {
		testName string
		query    url.Values
	}{
		{"empty", make(url.Values)},
		{"one", url.Values{"flavor": []string{"sour"}}},
		{"two", url.Values{"treats": []string{"gelato", "cake"}}},
		{"multi", url.Values{"treats": []string{"gelato", "cake"}, "names": []string{"diesel", "widow"}}},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			actualQuery := make(url.Values)
			req := olives.NewRequest()
			for name, vals := range tt.query {
				for _, val := range vals {
					req = req.WithQuery(name, val)
				}
			}

			w, err := req.Send(handler(actualQuery))
			if err != nil {
				t.Errorf("Unexpected olives error: %s", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Got unexpected response status: %d", w.Code)
			}

			mapSlicesEq(t, tt.query, actualQuery)
		})
	}
}

func TestPath(t *testing.T) {
	handler := func(vals map[string]string) func(c *gin.Context) {
		return func(c *gin.Context) {
			for _, p := range c.Params {
				vals[p.Key] = p.Value
			}

			c.Status(http.StatusOK)
		}
	}

	testCases := []struct {
		testName   string
		pathParams map[string]string
	}{
		{"empty", make(map[string]string)},
		{"one", map[string]string{"flavor": "sour"}},
		{"multi", map[string]string{"cookies": "girl-scout", "gelato": "grape"}},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			actualValues := make(map[string]string)

			req := olives.NewRequest()
			for name, val := range tt.pathParams {
				req = req.WithPath(name, val)
			}

			w, err := req.Send(handler(actualValues))
			if err != nil {
				t.Errorf("Unexpected olives error: %s", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Got unexpected response status: %d", w.Code)
			}

			mapsEq(t, tt.pathParams, actualValues)
		})
	}
}

func TestBody(t *testing.T) {
}

func mapSlicesEq(t *testing.T, exp, act map[string][]string) bool {
	if len(exp) != len(act) {
		t.Errorf("Maps had different sizes. Expected:\n%v (%d)\nActual: %v (%d)", exp, len(exp), act, len(act))
		return false
	}

EXP:
	for i, e := range exp {
		for j, a := range act {
			if i == j {
				if !slicesEq(t, e, a) {
					t.Errorf("Key %s had mis-matches slices", i)
				}
				delete(act, j)
				delete(exp, i)
				goto EXP
			}
		}
	}

	if len(act) == 0 {
		return true
	}

	t.Errorf("Maps had different keys. Expected, but not found:\n%v\nFound, but not expected:\n%v", exp, act)
	return false
}

func mapsEq(t *testing.T, exp, act map[string]string) bool {
	if len(exp) != len(act) {
		t.Errorf("Maps had different sizes. Expected:\n%v (%d)\nActual: %v (%d)", exp, len(exp), act, len(act))
		return false
	}

EXP:
	for i, e := range exp {
		for j, a := range act {
			if i == j {
				if a != e {
					t.Errorf("Expected '%s' for key '%s', but found '%s'", e, i, a)
				}
				delete(act, j)
				delete(exp, i)
				goto EXP
			}
		}
	}

	if len(act) == 0 {
		return true
	}

	t.Errorf("Maps had different keys. Expected, but not found:\n%v\nFound, but not expected:\n%v", exp, act)
	return false
}

func slicesEq(t *testing.T, exp, act []string) bool {
	if len(exp) != len(act) {
		t.Errorf("Slices had different lengths. Expected:\n%v (%d)\nActual: %v (%d)", exp, len(exp), act, len(act))
		return false
	}

EXP:
	for i, e := range exp {
		for j, a := range act {
			if e == a {
				act = append(act[:j], act[j+1:]...)
				exp = append(exp[:i], exp[i+1:]...)
				goto EXP
			}
		}
	}

	if len(act) == 0 {
		return true
	}

	t.Errorf("Slices had different elements. Expected, but not found:\n%v\nFound, but not expected:\n%v", exp, act)
	return false
}
