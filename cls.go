package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Net int

const (
	InNet Net = iota
	OutNet
)

type Client struct {
	client *http.Client

	Region string
	Host   string

	AuthParam

	common service

	Log *LogService
}

type AuthParam struct {
	SecretId  string
	SecretKey string
	Token     string

	// dynamic set backend cls host
	Host string

	// dynamic set cls header host
	HeaderHost string
}

type service struct {
	client *Client
}

func NewClient(region, sid, skey, token string, net Net) *Client {
	c := &Client{
		Region: region,
		AuthParam: AuthParam{
			SecretId: sid, SecretKey: skey, Token: token,
		},
	}
	c.common.client = c
	switch net {
	case InNet:
		c.Host = fmt.Sprintf("%s%s", region, ".cls.tencentyun.com")
	case OutNet:
		c.Host = fmt.Sprintf("%s%s", region, ".cls.tencentcs.com")
	default:
		c.Host = fmt.Sprintf("%s%s", region, ".cls.tencentcs.com")
	}
	c.client = http.DefaultClient
	c.Log = (*LogService)(&c.common)
	return c
}

func (c *Client) SetHost(host string) {
	if host == "" {
		return
	}
	c.Host = host
}

func (c *Client) WithHttpClient(httpc *http.Client) *Client {
	c.client = httpc
	return c
}

func (c *Client) NewRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		switch body.(type) {
		// only for pb binary data
		case []byte:
			buf = bytes.NewBuffer(body.([]byte))
		default:
			buf = new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			err := enc.Encode(body)
			if err != nil {
				return nil, err
			}
		}
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = withContext(ctx, req)
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

// See https://cloud.tencent.com/document/product/614/12402
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		respData := new(ErrorResponse)
		if err = json.Unmarshal(data, respData); err != nil {
			errorResponse.ErrorCode = "00500"
			errorResponse.ErrorMessage = fmt.Sprintf("resp data not json format, json.Unmarshal error, data: %s", data)
		} else {
			errorResponse.ErrorCode = respData.ErrorCode
			errorResponse.ErrorMessage = respData.ErrorMessage
		}
	}
	return errorResponse
}

func SetAuthorizationHeader(req *http.Request, sid, skey, token, method, uri string, params, headers map[string]string) {
	// Step 1. Gen SignKey
	keyTimeStart := time.Now().Unix()
	keyTimeEnd := keyTimeStart + 30*3600 // 30 minute
	KeyTime := fmt.Sprintf("%d;%d", keyTimeStart, keyTimeEnd)
	mac := hmac.New(sha1.New, []byte(skey))
	mac.Write([]byte(KeyTime))
	SignKey := fmt.Sprintf("%x", mac.Sum(nil))

	// Step 2. Gen FormatString
	FormatMethod := strings.ToLower(method)
	FormatParams := ""
	paramList := ""
	if len(params) > 0 {
		paramsKeys := []string{}
		for key := range params {
			paramsKeys = append(paramsKeys, key)
		}
		sort.Strings(paramsKeys)

		for _, v := range paramsKeys {
			paramList += strings.ToLower(v) + ";"
			FormatParams += strings.ToLower(url.QueryEscape(v)) + "=" + strings.ToLower(url.QueryEscape(params[v])) + "&"
		}
		if len(FormatParams) > 0 {
			paramList = paramList[:len(paramList)-1]
			FormatParams = FormatParams[:len(FormatParams)-1]
		}
	}

	FormatHeaders := ""
	headerList := ""
	if len(headers) > 0 {
		headerKeys := []string{}
		for key := range headers {
			headerKeys = append(headerKeys, key)
		}
		sort.Strings(headerKeys)

		for _, v := range headerKeys {
			headerList += strings.ToLower(v) + ";"
			FormatHeaders += strings.ToLower(v) + "=" + url.QueryEscape(headers[v]) + "&"
		}
		if len(headerList) > 0 {
			headerList = headerList[:len(headerList)-1]
			FormatHeaders = FormatHeaders[:len(FormatHeaders)-1]
		}
	}
	FormatString := fmt.Sprintf("%s\n%s\n%s\n%s\n", FormatMethod, uri, FormatParams, FormatHeaders)

	// Step 3. Gen StringToSign
	h := sha1.New()
	io.WriteString(h, FormatString)
	SHA1HashFormatString := fmt.Sprintf("%x", h.Sum(nil))
	SignAlgorithm := "sha1"
	// SignKey 和 StringToSign 使用相同的有效起止时间

	SignTime := KeyTime
	StringToSign := fmt.Sprintf("%s\n%s\n%s\n", SignAlgorithm, SignTime, SHA1HashFormatString)

	// Step 4. Gen Signature
	mac = hmac.New(sha1.New, []byte(SignKey))
	mac.Write([]byte(StringToSign))
	Signature := fmt.Sprintf("%x", mac.Sum(nil))

	// Step 5. Gen Authorization
	Authorization := fmt.Sprintf("q-sign-algorithm=%s&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=%s&q-url-param-list=%s&q-signature=%s",
		SignAlgorithm, sid, SignTime, KeyTime, headerList, paramList, Signature)
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("X-Cls-Token", token)
	req.Host = headers["Host"]
}

func withContext(ctx context.Context, req *http.Request) *http.Request {
	return req.WithContext(ctx)
}

func (s *service) AuthParam() AuthParam {
	return AuthParam{
		SecretId:  s.client.SecretId,
		SecretKey: s.client.SecretKey,
		Token:     s.client.Token,
	}
}

type LogService service

// S represent success
type S struct{}

func (s *LogService) UploadStructuredLog(ctx context.Context, topicId string, body []byte) (*S, *http.Response, error) {
	return s.UploadStructuredLogWithAuthParam(ctx, topicId, body, (*service)(s).AuthParam())
}

func (s *LogService) UploadStructuredLogWithAuthParam(ctx context.Context, topicId string, body []byte, authParam AuthParam) (*S, *http.Response, error) {
	var (
		uri      = "/structuredlog"
		method   = "POST"
		paramStr = fmt.Sprintf("?topic_id=%s", topicId)
		url      = fmt.Sprintf("http://%s%s%s", s.client.Host, uri, paramStr)

		params  = map[string]string{"topic_id": topicId}
		headers = map[string]string{"Host": s.client.Host}
	)

	if authParam.Host != "" {
		url = fmt.Sprintf("http://%s%s%s", authParam.Host, uri, paramStr)
	}

	if authParam.HeaderHost != "" {
		headers["Host"] = authParam.HeaderHost
	}

	req, err := s.client.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}
	SetAuthorizationHeader(
		req, authParam.SecretId, authParam.SecretKey, authParam.Token,
		method, uri, params, headers,
	)
	req.Header.Set("Content-Type", "application/x-protobuf")
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, resp, err
	}

	return &S{}, resp, nil
}

func (s *LogService) UploadStructuredLogUsingJsonWithAuthParam(ctx context.Context, topicId string, body []byte, authParam AuthParam) (*S, *http.Response, error) {
	var (
		uri      = "/structuredlog"
		method   = "POST"
		paramStr = fmt.Sprintf("?topic_id=%s", topicId)
		url      = fmt.Sprintf("http://%s%s%s", s.client.Host, uri, paramStr)

		params  = map[string]string{"topic_id": topicId}
		headers = map[string]string{"Host": s.client.Host}
	)

	if authParam.Host != "" {
		url = fmt.Sprintf("http://%s%s%s", authParam.Host, uri, paramStr)
	}

	if authParam.HeaderHost != "" {
		headers["Host"] = authParam.HeaderHost
	}

	req, err := s.client.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}
	SetAuthorizationHeader(
		req, authParam.SecretId, authParam.SecretKey, authParam.Token,
		method, uri, params, headers,
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, resp, err
	}

	return &S{}, resp, nil
}
