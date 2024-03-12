package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	cli *http.Client
)

func init() {
	cli = &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: time.Second * 3,
			}).DialContext,
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          500,
			IdleConnTimeout:       time.Second * 30,
			ExpectContinueTimeout: time.Second,
		},
	}
}

type Request struct {
	URL    string
	Header http.Header
	Params url.Values
	Form   url.Values
	Body   any
}

// Call 发起 http 请求
func Call(ctx context.Context, req *Request) (*Response, error) {
	var httpReq *http.Request
	var err error
	if req.Body != nil {
		bys, _ := json.Marshal(req.Body)
		bodyBuf := bytes.NewBuffer(bys)
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, req.URL, bodyBuf)
	} else if len(req.Form) > 0 {
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, req.URL, strings.NewReader(req.Form.Encode()))
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	}
	if err != nil {
		return nil, err
	}
	if len(req.Params) > 0 {
		httpReq.URL.RawQuery = req.Params.Encode()
	}
	return doRequest(httpReq)
}

func doRequest(req *http.Request) (*Response, error) {
	start := time.Now()
	resp, err := cli.Do(req)
	response := &Response{
		sendAt:      start,
		RawResponse: resp,
	}
	if err != nil {
		return response, err
	}
	defer closeq(resp.Body)
	body := resp.Body
	if response.body, err = io.ReadAll(body); err != nil {
		return response, err
	}
	response.size = int64(len(response.body))
	return response, nil
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		silently(c.Close())
	}
}

func silently(_ ...interface{}) {}

type Response struct {
	RawResponse *http.Response

	body       []byte
	size       int64
	sendAt     time.Time
	receivedAt time.Time
}

func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.body
}

func (r *Response) Status() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

// Proto method returns the HTTP response protocol used for the request.
func (r *Response) Proto() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

// Header method returns the response headers
func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

// Cookies method to access all the response cookies
func (r *Response) Cookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}
	return r.RawResponse.Cookies()
}

// String method returns the body of the server response as String.
func (r *Response) String() string {
	if r.body == nil {
		return ""
	}
	return strings.TrimSpace(string(r.body))
}

func (r *Response) Size() int64 {
	return r.size
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}

// FmtBody body bytes 转对象
func (r *Response) FmtBody(model interface{}) error {
	return json.Unmarshal(r.Body(), model)
}
