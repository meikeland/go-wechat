package main

import (
	"log"
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meikeland/go-wechat/wechat"
)

var (
	wechatClient wechat.ApiClient
)

func main() {
	wechatClient := wechat.New("your app id", "your app secret")
	accessToken, _ := wechatClient.GetAccessToken()
	log.Print(accessToken)

	router := gin.Default()
	router.Use(wechatOAuth)
	router.GET("landing", landingFromWeChat)
}

func wechatOAuth(c *gin.Context) {
	reqPath := c.Request.URL.Path
	if strings.Contains(reqPath, "/landing") {
		c.Next()
		return
	}

	openID, _ := c.Request.Cookie("openID")
	if openID == nil {
		oauthURL := wechatClient.OAuth.BuildOAuthPage("http://www.meistore.cn/landing", c.Request.RequestURI)
		c.Redirect(http.StatusTemporaryRedirect, oauthURL)
		c.Abort()
	} else {
		c.Set("openID", openID.Value)
		c.Next()
	}
}

func landingFromWeChat(c *gin.Context) {
	code := c.Query("code")
	from := c.Query("from")
	user, err := wechatClient.OAuth.GetUserByCode(code)
	if err != nil {
		panic(err.Error())
	}

	log.Print(user)
	log.Print(from)
}
