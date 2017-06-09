package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
)

// Wechat 的所有变量
type Wechat struct {
	clientMu sync.Mutex   // clientMu protects the client during calls that modify the CheckRedirect func.
	client   *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public GitHub API, but can be
	// set to a domain endpoint to use with GitHub Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	appKey    string
	appSecret string

	tokenSwitch bool // true 开启access token 维护机制, false 关闭
	accessToken string

	common service

	User  *UserService
	Card  *CardService
	Token *TokenService
}

type service struct {
	wechat *Wechat
}

// New 生成一个wechat实例
func New(appkey, appSecret string) *Wechat {
	w := &Wechat{appKey: appkey, appSecret: appSecret}
	w.tokenSwitch = false
	w.User = (*UserService)(&w.common)
	w.Card = (*CardService)(&w.common)
	w.Token = (*TokenService)(&w.common)
	return w
}

// StartToken 开始token维护工作
func (w *Wechat) StartToken() *Wechat {
	w.tokenSwitch = true
	w.Token.Start()
	return w
}

// Key 获取app key
func (w *Wechat) Key() string {
	return w.appKey
}

// Secret 获取app secret
func (w *Wechat) Secret() string {
	return w.appSecret
}

// GetAccessToken 获取 access token
func (w *Wechat) GetAccessToken() (string, error) {
	if !w.tokenSwitch {
		return "", errors.Errorf("未开启token自动维护机制")
	}
	return w.accessToken, nil
}

// saveAccessToken 保存 access token
func (w *Wechat) saveAccessToken(accessToken string) {
	w.accessToken = accessToken
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (w *Wechat) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := w.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it. If rate limit is exceeded and reset time is in the future,
// Do returns *RateLimitError immediately without making a network API call.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
func (w *Wechat) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = url.String()
				return nil, e
			}
		}

		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}
