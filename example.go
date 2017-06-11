package main

import (
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meikeland/go-wechat/wechat"
)

var (
	wechatClient *wechat.ApiClient
)

func main() {
	wechatClient = wechat.New("wxadbd9d3df031fabf", "3b4993c244c39759727c863de53baddb")

	router := gin.Default()
	router.Use(wechatOAuth)
	router.GET("/", index)
	router.GET("/landing", landingFromWeChat)
	log.Fatal(router.Run(":80"))
}

// gin的middleware，用于验证处理微信网页授权
func wechatOAuth(c *gin.Context) {
	reqPath := c.Request.URL.Path
	if strings.Contains(reqPath, "/landing") {
		c.Next()
		return
	}

	openID := c.Query("openID")
	if len(openID) == 0 {
		oauthURL := wechatClient.OAuth.Link("http://www.meistore.cn/landing", c.Request.RequestURI)
		c.Redirect(http.StatusTemporaryRedirect, oauthURL)
		c.Abort()
	} else {
		c.Set("openID", openID)
		c.Next()
	}
}

// index 首页，如果已经授权过，则会拿到openID
func index(c *gin.Context) {
	openID, _ := c.Get("openID")
	c.String(200, "你的openID是%s", openID)
}

// landingFromWeChat 从微信授权后回到本站点的页面，自带code和from参数
func landingFromWeChat(c *gin.Context) {
	code := c.Query("code")
	from := c.Query("from")
	user, err := wechatClient.OAuth.GetUserByCode(code)
	if err != nil {
		panic(err.Error())
	}

	// 为演示方便，这里直接将获取到的openID放到请求参数里了
	from = fmt.Sprintf("%s?openID=%s", from, user.OpenID)
	c.Redirect(http.StatusFound, from)
}
