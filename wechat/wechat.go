package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gotit/errors"
)

const (
	// CachePolicyNone 缓存AccessToken的策略为不缓存
	CachePolicyNone = "none"
	// CachePolicyAutonomy 缓存AccessToken的策略为自治，自主进行7200秒的刷新
	CachePolicyAutonomy = "autonomy"
	// CachePolicyHTTP 缓存AccessToken的策略为http，不自己维护AccessToken，每次需要，发送http请求到中控服务器获取
	CachePolicyHTTP = "http"
)

// APIConfig 调用微信Api的配置参数
type APIConfig struct {
	AppID                   string // 公众号AppID
	AppSecret               string // 公众号AppSecret
	MchID                   string // 商户ID
	MchSecret               string // 商户Secret
	MemberCardID            string // 会员卡ID
	AccessTokenCachePolicy  string // 公众号AccessToken缓存策略
	AccessTokenCacheAddress string // 公众号AccessToken缓存地址，当APIClient需要用到AccessToken时，会去这个地址获取，只有在policy是http的情况下有用到
}

// APIClient 的所有变量
type APIClient struct {
	client                  *http.Client // HTTP client used to communicate with the API.
	BaseURL                 *url.URL
	AppID                   string              // 公众号AppID
	AppSecret               string              // 公众号AppSecret
	MchID                   string              // 商户ID
	MchSecret               string              // 商户Secret
	MemberCardID            string              // 会员卡ID
	accessTokenCachePolicy  string              // 公众号AccessToken缓存策略
	accessTokenCacheAddress string              // 公众号AccessToken缓存地址
	autonomicAccessToken    string              // 当accessToken缓存策略为自治时，动态的维护这个accessToken，其他策略下，该数值为空
	common                  service             // Reuse a single struct instead of allocating one for each service on the heap.
	User                    *UserService        // 与微信公众平台服务的用户管理相关接口
	Card                    *CardService        // 与微信公众平台服务的微信卡券相关接口
	Pay                     *PayService         // 与微信商户平台服务的微信支付相关接口
	AccessToken             *AccessTokenService // 与微信公众平台服务的AccessToken相关接口
	OAuth                   *OAuthService       // 与微信公众平台服务的网页授权相关接口
}

type service struct {
	wechat *APIClient
}

// New 生成一个wechat实例
func New(config *APIConfig) *APIClient {
	w := &APIClient{
		client:                 http.DefaultClient,
		AppID:                  config.AppID,
		AppSecret:              config.AppSecret,
		MchID:                  config.MchID,
		MchSecret:              config.MchSecret,
		MemberCardID:           config.MemberCardID,
		accessTokenCachePolicy: config.AccessTokenCachePolicy,
	}

	w.common.wechat = w

	w.User = (*UserService)(&w.common)
	w.Card = (*CardService)(&w.common)
	w.Pay = (*PayService)(&w.common)
	w.AccessToken = (*AccessTokenService)(&w.common)
	w.OAuth = (*OAuthService)(&w.common)

	// 根据AccessToken缓存机制的设置进行初始化
	switch w.accessTokenCachePolicy {
	case CachePolicyNone:
		log.Print("不维护AccessToken")
	case CachePolicyAutonomy:
		// 开始自治维护AccessToken
		w.startTimer()
		log.Print("自治维护AccessToken")
	case CachePolicyHTTP:
		if strings.HasPrefix(config.AccessTokenCacheAddress, "http://") {
			// 使用AccessTokenCacheAddress
			w.accessTokenCacheAddress = config.AccessTokenCacheAddress
		} else {
			panic("缓存AccessToken的中控服务器设置无效")
		}
		log.Print("用中控方式获取AccessToken")
	default:
		panic("缓存access_token的机制未正确设置")
	}
	return w
}

// GetAccessToken 获取 access token
func (w *APIClient) GetAccessToken() (string, error) {
	switch w.accessTokenCachePolicy {
	case CachePolicyAutonomy:
		return w.autonomicAccessToken, nil
	case CachePolicyHTTP:
		return "暂未实现", errors.New("暂未实现")
	default:
		return "", errors.New("无法确定的access_token缓存设置")
	}
}

// NewRequest 创建一个api请求体, 以json发送body参数
func (w *APIClient) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
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
func (w *APIClient) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
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
			body, err := ioutil.ReadAll(resp.Body)
			log.Printf("url %s body %s", req.URL.Path, string(body))
			err = json.NewDecoder(resp.Body).Decode(v)

			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}

// startTimer 启动自治维护AccessToken的timer
func (w *APIClient) startTimer() {
	if !w.AccessToken.Verify(w.autonomicAccessToken) {
		accessToken, _, err := w.AccessToken.Get()
		if err != nil {
			log.Print(err)
		} else {
			w.autonomicAccessToken = accessToken
			log.Printf("最新获取到的AccessToken是: %s", accessToken)
		}
	}

	time.AfterFunc(durationRefreshAccessToken, func() {
		w.startTimer()
	})
}
