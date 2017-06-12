package wechat

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// AccessTokenService 微信 access token 的获取、维护服务
type AccessTokenService service

const (
	durationRefreshAccessToken = time.Second * 30
	urlGetAccessToken          = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	urlVerifyAccessToken       = "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=%s"
)

// Verify 验证当前的access_token是否有效，由于微信并没有提供一个验证有效性的接口
// 所以实际的验证算法是利用了一个微信公众号可无限次调用的接口，来判断是否access_token已失效
// 这个接口是获取微信服务器IP地址，在微信公众号文档的"开始开发"章节
// 这个接口可能会随着微信的功能改进而取消掉，但至少在目前是可用的
func (s *AccessTokenService) Verify(accessToken string) bool {
	url := fmt.Sprintf(urlVerifyAccessToken, accessToken)

	req, err := s.wechat.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("获取AccessToken请求异常 error: %s", err.Error())
		return false
	}

	result := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	_, err = s.wechat.Do(nil, req, result)
	if err != nil {
		return false
	}
	if result.Errcode != 0 {
		return false
	}

	return true
}

// Get 从微信服务器获取AccessToken
func (s *AccessTokenService) Get() (string, int64, error) {
	url := fmt.Sprintf(urlGetAccessToken, s.wechat.AppID, s.wechat.AppSecret)
	req, err := s.wechat.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("获取AccessToken请求异常 error: %s", err.Error())
		return "", 0, errors.New("无法创建获取AccessToken的请求")
	}

	token := &struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Errorcode    int    `json:"errcode"`
		ErrorMessage string `json:"errmsg"`
	}{}
	_, err = s.wechat.Do(nil, req, token)
	if err != nil {
		log.Printf("获取AccessToken请求异常 error: %s", err.Error())
		return "", 0, errors.New("无法获取AccessToken的请求")
	}

	if token.Errorcode != 0 {
		log.Print(token.ErrorMessage)
		return "", 0, errors.New(token.ErrorMessage)
	}

	if len(token.AccessToken) == 0 {
		log.Println("获取AccessToken失败，未知原因")
	}
	return token.AccessToken, token.ExpiresIn, nil
}
