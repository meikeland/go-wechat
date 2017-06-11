package wechat

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	urlGetOAuthAccessToken = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	urlGetOAuthUserInfo    = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN"
	urlOAuthPage           = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo#wechat_redirect"
)

// OAuthUser 通过微信网页授权拉取的用户信息
// 此用户信息只有在scope为snsapi_userinfo时
// 通过access_token和openid拉取
type OAuthUser struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"` //用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
}

// OAuthService 在微信客户端中访问第三方网页，利用微信网页授权机制，
// 来获取用户基本信息
type OAuthService service

// oauthAccessToken 微信网页授权时使用的access_token，与公众号基础的access_token不同
type oauthAccessToken struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	OpenID       string  `json:"openid"`
	UnionID      string  `json:"unionid"`
	Scope        string  `json:"scope"`
}

// GetUserByCode 直接通过code获取OAuthUser
func (s *OAuthService) GetUserByCode(code string) (*OAuthUser, error) {
	// 第一步，用code从微信服务器换取access_token
	url := fmt.Sprintf(urlGetOAuthAccessToken, s.wechat.AppID, s.wechat.AppSecret, code)
	req, err := s.wechat.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	token := &oauthAccessToken{}
	_, err = s.wechat.Do(nil, req, token)
	if err != nil {
		return nil, err
	}
	if len(token.AccessToken) == 0 {
		return nil, errors.New("通过code获取OAuthAccessToken失败")
	}

	// 第二步，用access_token去拉取用户信息
	return s.GetUserByAccessToken(token.AccessToken, token.OpenID)
}

// GetUserByAccessToken 当scope为snsapi_userinfo时，通过access_token和openid拉取用户信息
func (s *OAuthService) GetUserByAccessToken(accessToken, openID string) (*OAuthUser, error) {
	url := fmt.Sprintf(urlGetOAuthUserInfo, accessToken, openID)
	req, err := s.wechat.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	oauthUser := &OAuthUser{}
	_, err = s.wechat.Do(nil, req, oauthUser)
	if err != nil {
		return nil, err
	}
	return oauthUser, err
}

// Link 生成微信网页授权的页面地址，这个地址中的redirect_uri参数
// 是在用户同意微信授权之后，重定向到第三方网站的页面，因此landingPage是一个形如: https://www.abc.com/login的地址
// 这个地址会接收到微信提供的code参数，如果想要回到授权前的页面，应传入from参数，
// 在正确使用code拉取用户信息之后，回到from这个页面
func (s *OAuthService) Link(landingPage, from string) string {
	// 将from转换为landingPage?from={from}的地址
	redirectURL := url.QueryEscape(fmt.Sprintf("%s?from=%s", landingPage, from))
	return fmt.Sprintf(urlOAuthPage, s.wechat.AppID, redirectURL)
}
