package easy_http

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type requestHandler struct {
	client         *http.Client
	http2          bool
	transport      *http.Transport
	proxy          func(*http.Request) (*url.URL, error)
	Url            string
	Headers        map[string][]string
	Cookies        map[string][]string
	Method         string
	Body           string
	Timeout        time.Duration
	FollowRedirect bool
	IgnoreTLS      bool
	log            *zap.Logger
}

/*
核心配置结构
*/
type RequestConfig struct {
	Http2          bool
	Headers        map[string][]string
	Cookies        map[string][]string
	Body           string
	Timeout        time.Duration
	FollowRedirect bool
	ProxyUrl       string
	IgnoreTLS      bool
}

func NewRequest(method Method, u string, config *RequestConfig) *requestHandler {
	log, _ := zap.NewDevelopment()
	var transport = &http.Transport{}

	if config.ProxyUrl != "" {
		transport.Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(config.ProxyUrl)
		}
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.IgnoreTLS}
	if config.Http2 {
		http2.ConfigureTransport(transport)
	}
	c := &http.Client{
		Transport: transport,
	}
	handler := &requestHandler{
		client:         c,
		Url:            u,
		Method:         method.String(),
		Headers:        make(map[string][]string),
		Cookies:        make(map[string][]string),
		Body:           config.Body,
		Timeout:        config.Timeout,
		FollowRedirect: config.FollowRedirect,
		IgnoreTLS:      config.IgnoreTLS,
		transport:      transport,
		log:            log,
	}
	if config.Headers != nil {
		handler.Headers = config.Headers
	}
	if config.Cookies != nil {
		handler.Cookies = config.Cookies
	}
	return handler
}

func NewGet(url string) *requestHandler {
	log, _ := zap.NewDevelopment()
	handler := &requestHandler{
		client:    &http.Client{},
		Url:       url,
		Method:    GET.String(),
		Headers:   make(map[string][]string),
		Cookies:   make(map[string][]string),
		transport: &http.Transport{},
		log:       log,
	}
	return handler
}

func NewPost(url string, body string) *requestHandler {
	log, _ := zap.NewDevelopment()
	handler := &requestHandler{
		client:    &http.Client{},
		Url:       url,
		Method:    POST.String(),
		Body:      body,
		Headers:   make(map[string][]string),
		Cookies:   make(map[string][]string),
		transport: &http.Transport{},
		log:       log,
	}
	return handler
}

func (h *requestHandler) SetProxy(proxyUrl string) {
	h.proxy = func(*http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	h.transport.Proxy = h.proxy
	h.client.Transport = h.transport
}

func (h *requestHandler) SkipTLSCheck(skip bool) {
	h.transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skip}
	h.client.Transport = h.transport
}

func (h *requestHandler) EnableHttp2(enable bool) {
	if enable {
		http2.ConfigureTransport(h.transport)
		h.client.Transport = h.transport
	}
}

func (h *requestHandler) Execute() (r *response, err error) {

	if strings.TrimSpace(h.Url) == "" {
		h.log.Error("no url specified", zap.String("url", h.Url))
		return nil, errors.New("no url specified")
	}
	if h.Method == "UNKONWN" {
		h.log.Error("no method specified", zap.String("method", h.Method))
		return nil, errors.New("no method specified")
	}
	req, err := http.NewRequest(h.Method, h.Url, strings.NewReader(h.Body))
	if err != nil {
		h.log.Error("cannot build request", zap.String("method", h.Method), zap.String("url", h.Url), zap.String("body", h.Body))
		return nil, err
	}
	if h.Headers != nil && len(h.Headers) > 0 {
		req.Header = h.Headers
	}
	if h.Cookies != nil && len(h.Cookies) > 0 {
		var cookieString string
		for cookieName := range h.Cookies {
			cookieString += fmt.Sprintf("%v=%v;", cookieName, h.Cookies[cookieName])
		}
		req.Header.Add("Cookie", cookieString)
	}
	if h.Timeout != 0 {
		h.client.Timeout = h.Timeout
	}
	if !h.FollowRedirect {
		h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	startTime := time.Now()
	resp, err := h.client.Do(req)
	endTime := time.Now()
	if err != nil {
		h.log.Error("send request error", zap.String("error", err.Error()))
		r = NewResponse(resp, endTime.Sub(startTime), h.log)
		return r, err
	}

	r = NewResponse(resp, endTime.Sub(startTime), h.log)
	defer resp.Body.Close()
	return r, nil
}

type response struct {
	Body       string
	Headers    map[string][]string
	Status     string
	StatusCode int
	Proto      string
	Duration   time.Duration
	log        *zap.Logger
}

func NewResponse(resp *http.Response, duration time.Duration, log *zap.Logger) *response {
	if resp == nil {
		return &response{
			Duration: duration,
			log:      log,
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error")
		return &response{
			Duration: duration,
			log:      log,
		}
	}
	r := &response{
		Body:       string(body),
		Headers:    resp.Header,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Duration:   duration,
		Proto:      resp.Proto,
		log:        log,
	}
	return r
}

func (resp *response) GetBodyByType(t interface{}) error {
	return json.Unmarshal([]byte(resp.Body), &t)
}

type Method int

const (
	GET Method = iota
	POST
	PUT
	HEAD
	DELETE
	PATCH
	OPTION
)

func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case HEAD:
		return "HEAD"
	case DELETE:
		return "DELETE"
	case PATCH:
		return "PATCH"
	case OPTION:
		return "OPTION"
	default:
		return "UNKONWN"
	}
}
