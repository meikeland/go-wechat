package wechat

import (
	"fmt"
	"log"
	"time"
)

// TokenService 微信 access token 的获取、维护服务
type TokenService service

const (
	duration             = time.Second * 30
	urlGetAccessToken    = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	urlVerifyAccessToken = "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=%s"
)

// Start 开始token 自动刷新维护机制
func (t *TokenService) Start() {
	if !t.verifyAccessToken() {
		t.requestAccessToken()
	}
	time.AfterFunc(duration, func() {
		t.Start()
	})
}

func (t *TokenService) verifyAccessToken() (valid bool) {
	valid = true
	token, _ := t.wechat.GetAccessToken()
	url := fmt.Sprintf(urlVerifyAccessToken, token)

	req, err := t.wechat.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("获取AccessToken请求异常 error: %s", err.Error())
		valid = false
		return
	}

	result := struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	_, err = t.wechat.Do(nil, req, result)
	if err != nil {
		valid = false
		return
	}
	if result.Errcode != 0 {
		valid = false
		return
	}
	return
}

func (t *TokenService) requestAccessToken() {
	url := fmt.Sprintf(urlGetAccessToken, t.wechat.AppID, t.wechat.AppSecret)
	req, err := t.wechat.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("获取AccessToken请求异常 error: %s", err.Error())
		return
	}

	token := struct {
		AccessToken string  `json:"access_token"`
		ExpiresIn   float64 `json:"expires_in"`
	}{}
	_, err = t.wechat.Do(nil, req, token)
	if err != nil {
		return
	}

	if len(token.AccessToken) == 0 {
		log.Println("获取AccessToken失败，未知原因")
	}
	t.wechat.saveAccessToken(token.AccessToken)
}
