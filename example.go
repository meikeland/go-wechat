package main

import (
	"log"

	"github.com/meikeland/go-wechat/wechat"
)

func main() {
	wechat := wechat.New("your app id", "your app secret")
	accessToken, _ := wechat.GetAccessToken()
	log.Print(accessToken)
}
