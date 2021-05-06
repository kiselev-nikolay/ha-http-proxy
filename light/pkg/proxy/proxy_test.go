package proxy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kiselev-nikolay/ha-http-proxy/light/pkg/proxy"
)

func makeMock(code int, header map[string]string, body string) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		res := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
		for k, v := range header {
			res.Header.Add(k, v)
		}
		return res, nil
	}
}

func assertBody(t *testing.T, expect string, got *http.Response) {
	b, _ := ioutil.ReadAll(got.Body)
	gotBody := strings.TrimSpace(string(b))
	if gotBody != expect {
		t.Errorf(`Got %v; must be %v`, gotBody, expect)
	}
}

func TestNoJSONBody(t *testing.T) {
	h := &proxy.Handler{DoRequest: makeMock(200, map[string]string{}, "")}
	ts := httptest.NewServer(h)
	defer ts.Close()
	res, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(""))
	if err != nil {
		t.Fail()
		return
	}
	assertBody(t, `{"errors":["json body decode error"]}`, res)
}

func TestErrResponses(t *testing.T) {
	tests := []struct {
		name   string
		mock   func(req *http.Request) (*http.Response, error)
		send   *proxy.Request
		expect string
	}{
		{
			"EmptyJSON",
			makeMock(200, map[string]string{}, ""),
			&proxy.Request{},
			`{"errors":["method is empty","url is empty"]}`,
		},
		{
			"EmptyMethod",
			makeMock(200, map[string]string{}, ""),
			&proxy.Request{RawURL: "http://test.com"},
			`{"errors":["method is empty"]}`,
		},
		{
			"EmptyUrl",
			makeMock(200, map[string]string{}, ""),
			&proxy.Request{Method: "GET"},
			`{"errors":["url is empty"]}`,
		},
		{
			"InvalidURL",
			makeMock(200, map[string]string{}, ""),
			&proxy.Request{Method: "GET", RawURL: ":// test.com"},
			`{"errors":["url is is invalid"]}`,
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%vRequest", tt.name)
		t.Run(testname, func(t *testing.T) {
			h := &proxy.Handler{DoRequest: tt.mock}
			ts := httptest.NewServer(h)
			defer ts.Close()
			body, _ := json.Marshal(tt.send)
			res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fail()
			}
			assertBody(t, tt.expect, res)
		})
	}
}

func assertRes(t *testing.T, expect *proxy.Response, got *http.Response) {
	gotRes := &proxy.Response{}
	json.NewDecoder(got.Body).Decode(gotRes)
	if len(gotRes.ID) != 32 {
		t.Errorf(`Got invalid id in response, got %s`, gotRes.ID)
	}
	gotRes.ID = ""
	if fmt.Sprint(expect) != fmt.Sprint(gotRes) {
		t.Errorf(`Got %+v; must be %+v`, gotRes, expect)
	}
}

func TestProxy(t *testing.T) {
	tests := []struct {
		name   string
		mock   func(req *http.Request) (*http.Response, error)
		send   *proxy.Request
		expect *proxy.Response
	}{
		{
			"Get",
			makeMock(200, map[string]string{}, ""),
			&proxy.Request{Method: "GET", RawURL: "http://test.com"},
			&proxy.Response{Status: 200},
		},
		{
			"GetHeaders",
			makeMock(201, map[string]string{"X-Data": "Test"}, ""),
			&proxy.Request{Method: "GET", RawURL: "http://test.com"},
			&proxy.Response{Status: 201, Headers: map[string]string{"X-Data": "Test"}},
		},
		{
			"Body",
			makeMock(201, map[string]string{}, "Test me now"),
			&proxy.Request{Method: "GET", RawURL: "http://test.com"},
			&proxy.Response{Status: 201, Length: 11},
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%vRequest", tt.name)
		t.Run(testname, func(t *testing.T) {
			h := &proxy.Handler{DoRequest: tt.mock}
			ts := httptest.NewServer(h)
			defer ts.Close()
			body, _ := json.Marshal(tt.send)
			res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fail()
			}
			assertRes(t, tt.expect, res)
		})
	}
}

func TestSendHeaders(t *testing.T) {
	h := &proxy.Handler{DoRequest: func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Data") != "Test" {
			t.Fail()
			return &http.Response{}, fmt.Errorf("X-Data not found")
		}
		res := &http.Response{
			StatusCode: 200,
		}
		return res, nil
	}}
	ts := httptest.NewServer(h)
	defer ts.Close()
	body, _ := json.Marshal(&proxy.Request{Method: "GET", RawURL: "http://test.com", Headers: map[string]string{"X-Data": "Test"}})
	http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
}

func TestServerError(t *testing.T) {
	h := &proxy.Handler{DoRequest: func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("Fail")
	}}
	ts := httptest.NewServer(h)
	defer ts.Close()
	body, _ := json.Marshal(&proxy.Request{Method: "GET", RawURL: "http://test.com"})
	res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fail()
	}
	assertBody(t, `{"errors":["request failed"]}`, res)
}

func TestTraceIdPresent(t *testing.T) {
	var traceID string
	h := &proxy.Handler{DoRequest: func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Hhp-Trace-Id") == "" {
			t.Fail()
		}
		traceID = req.Header.Get("X-Hhp-Trace-Id")
		return &http.Response{StatusCode: 200}, nil
	}}
	ts := httptest.NewServer(h)
	defer ts.Close()
	body, _ := json.Marshal(&proxy.Request{Method: "GET", RawURL: "http://test.com"})
	rawRes, _ := http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
	res := &proxy.Response{}
	json.NewDecoder(rawRes.Body).Decode(res)
	if len(res.ID) != 32 {
		t.Errorf(`Got invalid id in response, got %s`, res.ID)
	}
	if traceID != res.ID {
		t.Errorf(`Got wrong id in response got %s; must %s`, res.ID, traceID)
	}
}
