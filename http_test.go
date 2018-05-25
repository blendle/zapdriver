package zapdriver_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/blendle/zapdriver"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHTTP(t *testing.T) {
	t.Parallel()

	req := &zapdriver.HTTPPayload{}
	field := zapdriver.HTTP(req)

	assert.Equal(t, zap.Object("httpRequest", req), field)
}

func TestNewHTTP(t *testing.T) {
	t.Parallel()

	var tests = map[string]struct {
		req  *http.Request
		res  *http.Response
		want *zapdriver.HTTPPayload
	}{
		"empty": {
			nil,
			nil,
			&zapdriver.HTTPPayload{},
		},

		"RequestMethod": {
			&http.Request{Method: "GET"},
			nil,
			&zapdriver.HTTPPayload{RequestMethod: "GET"},
		},

		"Status": {
			nil,
			&http.Response{StatusCode: 404},
			&zapdriver.HTTPPayload{Status: 404},
		},

		"UserAgent": {
			&http.Request{Header: http.Header{"User-Agent": []string{"hello world"}}},
			nil,
			&zapdriver.HTTPPayload{UserAgent: "hello world"},
		},

		"RemoteIP": {
			&http.Request{RemoteAddr: "127.0.0.1"},
			nil,
			&zapdriver.HTTPPayload{RemoteIP: "127.0.0.1"},
		},

		"Referrer": {
			&http.Request{Header: http.Header{"Referer": []string{"hello universe"}}},
			nil,
			&zapdriver.HTTPPayload{Referer: "hello universe"},
		},

		"Protocol": {
			&http.Request{Proto: "HTTP/1.1"},
			nil,
			&zapdriver.HTTPPayload{Protocol: "HTTP/1.1"},
		},

		"RequestURL": {
			&http.Request{URL: &url.URL{Host: "example.com", Scheme: "https"}},
			nil,
			&zapdriver.HTTPPayload{RequestURL: "https://example.com"},
		},

		"RequestSize": {
			&http.Request{Body: ioutil.NopCloser(strings.NewReader("12345"))},
			nil,
			&zapdriver.HTTPPayload{RequestSize: "5"},
		},

		"ResponseSize": {
			nil,
			&http.Response{Body: ioutil.NopCloser(strings.NewReader("12345"))},
			&zapdriver.HTTPPayload{ResponseSize: "5"},
		},

		"simple request": {
			httptest.NewRequest("POST", "/", strings.NewReader("12345")),
			nil,
			&zapdriver.HTTPPayload{
				RequestSize:   "5",
				RequestMethod: "POST",
				RemoteIP:      "192.0.2.1:1234",
				Protocol:      "HTTP/1.1",
				RequestURL:    "/",
			},
		},

		"simple response": {
			nil,
			&http.Response{Body: ioutil.NopCloser(strings.NewReader("12345")), StatusCode: 404},
			&zapdriver.HTTPPayload{ResponseSize: "5", Status: 404},
		},

		"request & response": {
			&http.Request{Method: "POST", Proto: "HTTP/1.1"},
			&http.Response{StatusCode: 200},
			&zapdriver.HTTPPayload{RequestMethod: "POST", Protocol: "HTTP/1.1", Status: 200},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, zapdriver.NewHTTP(tt.req, tt.res))
		})
	}
}
