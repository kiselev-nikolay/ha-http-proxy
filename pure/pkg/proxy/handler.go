package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kiselev-nikolay/ha-http-proxy/pure/pkg/trace"
)

type Handler struct {
	DoRequest func(req *http.Request) (*http.Response, error)
	Logger    Logger
	Traffic   *Traffic
}

func (h *Handler) sendOK(resJSON *json.Encoder, req *http.Request, response *Response) {
	resJSON.Encode(response)
	if h.Logger != nil {
		h.Logger.Printf("%v | %v | OK: %v", req.Method, req.RemoteAddr, response.ID)
	}
}

func (h *Handler) sendErr(resJSON *json.Encoder, req *http.Request, errors []string) {
	errorRes := &ErrResponse{
		Errors: errors,
	}
	resJSON.Encode(errorRes)
	if h.Logger != nil {
		h.Logger.Printf("%v | %v | Fail: %v", req.Method, req.RemoteAddr, strings.Join(errors, ","))
	}
}

func NewHandler(ctx *Context, client *http.Client) *Handler {
	return &Handler{
		DoRequest: client.Do,
		Logger:    ctx.Logger,
		Traffic:   ctx.Traffic,
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	resJSON := json.NewEncoder(rw)
	reqJSON := json.NewDecoder(req.Body)

	request := &Request{}
	err := reqJSON.Decode(request)
	if err != nil {
		rw.WriteHeader(400)
		h.sendErr(resJSON, req, []string{"json body decode error"})
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
		h.sendErr(resJSON, req, validationErrors)
		return
	}

	httpReq, err := http.NewRequest(request.Method, request.URL.String(), bytes.NewBufferString(""))
	if err != nil {
		rw.WriteHeader(400)
		h.sendErr(resJSON, req, []string{`request is is invalid`})
		return
	}

	for k, v := range request.Headers {
		httpReq.Header.Add(k, v)
	}
	traceID := trace.GenerateID()
	httpReq.Header.Add("X-Hhp-Trace-Id", traceID)
	res, err := h.DoRequest(httpReq)
	if err != nil {
		rw.WriteHeader(502)
		h.sendErr(resJSON, req, []string{"request failed"})
		return
	}

	length := 0
	if res.Body != nil {
		responseBody, err := ioutil.ReadAll(res.Body)
		if err == nil {
			length = len(responseBody)
		}
	}

	response := &Response{
		ID:      traceID,
		Status:  res.StatusCode,
		Length:  length,
		Headers: make(map[string]string),
	}
	for k, v := range res.Header {
		response.Headers[k] = v[0]
	}
	rw.WriteHeader(200)
	h.sendOK(resJSON, req, response)
	if h.Traffic != nil {
		h.Traffic.Records[traceID] = TrafficRecord{
			*request,
			*response,
		}
	}
}
