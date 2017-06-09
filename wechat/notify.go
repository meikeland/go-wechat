package wechat

import (
	"crypto/sha1"
	"fmt"
	"io"
	"sort"
	"strings"
)

/*
CheckNotify 检查微信通知token是否合法
	1、然后对token、timestamp、nonce三个参数进行字典序排序，并拼接排序结果
	2、对2步骤生成的结果进行sha1加密
	3、判断加密结果与signature的一致性*/
func CheckNotify(signature, timestamp, nonce, token string) bool {
	strs := []string{token, timestamp, nonce}
	sort.Strings(strs)
	result := ""
	for _, s := range strs {
		result = fmt.Sprintf("%s%s", result, s)
	}
	t := sha1.New()
	io.WriteString(t, result)
	return strings.EqualFold(signature, fmt.Sprintf("%x", t.Sum(nil)))
}
