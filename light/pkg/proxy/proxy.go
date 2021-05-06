package proxy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/kiselev-nikolay/ha-http-proxy/light/pkg/trace"
)

func Run(addr string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	handler := &Handler{
		DoRequest: client.Do,
	}
	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	return err
}

type Handler struct {
	DoRequest func(req *http.Request) (*http.Response, error)
}

type Request struct {
	Method  string            `json:"method"`
	RawURL  string            `json:"url"`
	URL     *url.URL          `json:"-"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Length  int               `json:"length"`
	Headers map[string]string `json:"headers"`
}

type ErrResponse struct {
	Errors []string `json:"errors"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	resJSON := json.NewEncoder(rw)
	reqJSON := json.NewDecoder(req.Body)

	request := &Request{}
	err := reqJSON.Decode(request)
	if err != nil {
		rw.WriteHeader(400)
		resJSON.Encode(&ErrResponse{
			Errors: []string{"json body decode error"},
		})
		return
	}

	var validationErrors = make([]string, 0)
	if request.Method == "" {
		validationErrors = append(validationErrors, `method is empty`)
	}
	if request.RawURL == "" {
		validationErrors = append(validationErrors, `url is empty`)
	} else {
		url, err := url.Parse(request.RawURL)
		if err != nil {
			validationErrors = append(validationErrors, `url is is invalid`)
		}
		request.URL = url
	}
	if len(validationErrors) > 0 {
		rw.WriteHeader(400)
		resJSON.Encode(&ErrResponse{
			Errors: validationErrors,
		})
		return
	}

	httpReq := &http.Request{
		Method: request.Method,
		URL:    request.URL,
		Header: make(http.Header),
	}
	for k, v := range request.Headers {
		httpReq.Header.Add(k, v)
	}
	res, err := h.DoRequest(httpReq)
	if err != nil {
		rw.WriteHeader(502)
		resJSON.Encode(&ErrResponse{
			Errors: []string{"request failed"},
		})
		return
	}

	length := 0
	responseBody, err := ioutil.ReadAll(res.Body)
	if err == nil {
		length = len(responseBody)
	}

	response := &Response{
		ID:      trace.GetID(),
		Status:  res.StatusCode,
		Length:  length,
		Headers: make(map[string]string),
	}
	for k, v := range res.Header {
		response.Headers[k] = v[0]
	}
	rw.WriteHeader(200)
	resJSON.Encode(response)
}
