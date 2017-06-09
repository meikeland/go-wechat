package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gotit/errors"
)

// Wechat 的所有变量
type Wechat struct {
	client  *http.Client // HTTP client used to communicate with the API.
	BaseURL *url.URL

	AppID       string
	AppSecret   string
	tokenSwitch bool // true 开启access token 维护机制, false 关闭
	accessToken string

	common service

	User  *UserService
	Card  *CardService
	Pay   *PayService
	Token *TokenService
	OAuth *OAuthService
}

type service struct {
	wechat *Wechat
}

// New 生成一个wechat实例
func New(appkey, appSecret string) *Wechat {
	w := &Wechat{client: http.DefaultClient, tokenSwitch: false, AppID: appkey, AppSecret: appSecret}
	w.User = (*UserService)(&w.common)
	w.Card = (*CardService)(&w.common)
	w.Pay = (*PayService)(&w.common)
	w.Token = (*TokenService)(&w.common)
	return w
}

// StartToken 开始token维护工作
func (w *Wechat) StartToken() *Wechat {
	w.tokenSwitch = true
	w.Token.Start()
	return w
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

// NewRequest 创建一个api请求体, 以json发送body参数
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

// Do 执行http请求，并默认用json解析返回数据到结构体v
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
