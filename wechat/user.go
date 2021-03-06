package wechat

import "fmt"

// UserService 处理与用户相关的API，包括用户授权登录和获取、更新用户资料
type UserService service

const (
	urlGetUserInfo = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"
)

// WXUserInfo 微信用户基本信息
type WXUserInfo struct {
	Subscribe     int     `json:"subscribe"`      // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，拉取不到其余信息。
	Openid        string  `json:"openid"`         // 用户的标识，对当前公众号唯一
	Nickname      string  `json:"nickname"`       // 用户的昵称
	Sex           int     `json:"sex"`            // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Language      string  `json:"language"`       // 用户的语言，简体中文为zh_CN
	City          string  `json:"city"`           // 用户所在城市
	Province      string  `json:"province"`       // 用户所在省份
	Country       string  `json:"country"`        // 用户所在国家
	Headimgurl    string  `json:"headimgurl"`     // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空。若用户更换头像，原有头像URL将失效。
	SubscribeTime int64   `json:"subscribe_time"` // 用户关注时间，为时间戳。如果用户曾多次关注，则取最后关注时间
	Unionid       string  `json:"unionid"`        // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
	Remark        string  `json:"remark"`         // 公众号运营者对粉丝的备注，公众号运营者可在微信公众平台用户管理界面对粉丝添加备注
	Groupid       int64   `json:"groupid"`        // 用户所在的分组ID（兼容旧的用户分组接口）
	TagidList     []int64 `json:"tagid_list"`     // 用户被打上的标签ID列表
}

// GetUserInfoByOpenid 通过openid获取用户基本信息
func (s *UserService) GetUserInfoByOpenid(openid string) (*WXUserInfo, error) {
	token, err := s.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(urlGetUserInfo, token, openid)
	req, err := s.wechat.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	user := &WXUserInfo{}
	_, err = s.wechat.Do(nil, req, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
