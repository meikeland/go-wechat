package wechat

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotit/errors"
)

// CardMemberInfo 微信会员卡的会员信息
type CardMemberInfo struct {
	CardID       string
	Openid       string
	Unionid      string
	CardCode     string
	Nickname     string
	MobileNumber string
	Gender       string
	RealName     string
	Birthday     time.Time
	MemberStatus string
	Baby1        string
	Baby2        string
}

// CardMemberField 卡会员的参数
type CardMemberField struct {
	Name  string `json:"name"`  // 会员信息类目名称
	Value string `json:"value"` // 会员卡信息类目值，比如等级值等
}

// GetMemberByCode 通过会员卡号获取会员信息
func (c *CardService) GetMemberByCode(cardCode string) (*CardMemberInfo, error) {
	token, err := c.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	// 根据cardID和卡号拉取会员信息
	url := fmt.Sprintf(urlGetCardUserInfo, token)
	param := map[string]interface{}{
		"card_id": c.wechat.MemberCardID,
		"code":    cardCode,
	}
	req, err := c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return nil, err
	}
	member := &struct {
		Errcode          int    `json:"errcode"`           // 错误码，0为正常
		Errmsg           string `json:"errmsg"`            // 错误信息
		Openid           string `json:"openid"`            // 用户在本公众号内唯一识别码
		Nickname         string `json:"nickname"`          // 用户昵称
		MembershipNumber string `json:"membership_number"` // 积分信息
		Bonus            int    `json:"bonus"`             // 余额信息
		Sex              string `json:"sex"`               // 用户性别
		UserInfo         struct {
			CommonFieldList []CardMemberField `json:"common_field_list"`
			CustomFieldList []CardMemberField `json:"custom_field_list"`
		} `json:"user_info"` // 会员信息
		UserCardStatus string `json:"user_card_status"` // 当前用户会员卡状态，NORMAL 正常 EXPIRE 已过期 GIFTING 转赠中 GIFT_SUCC 转赠成功 GIFT_TIMEOUT 转赠超时 DELETE 已删除，UNAVAILABLE 已失效
		HasActive      bool   `json:"has_active"`       // 当前用户会员卡是否已激活
	}{}
	_, err = c.wechat.Do(nil, req, member)
	if err != nil {
		return nil, err
	}
	if !member.HasActive {
		return nil, errors.New("用户还未激活会员卡")
	}

	mobileNumber, realname, gender, birthday := unmarshalCardMemberFields(member.UserInfo.CommonFieldList)
	baby1, baby2 := unmarshalCardMemberCustomFields(member.UserInfo.CustomFieldList)
	info := &CardMemberInfo{
		CardID:       c.wechat.MemberCardID,
		Openid:       member.Openid,
		Unionid:      "",
		CardCode:     cardCode,
		Nickname:     member.Nickname,
		MobileNumber: mobileNumber,
		Gender:       gender,
		RealName:     realname,
		Birthday:     birthday,
		MemberStatus: member.UserCardStatus,
		Baby1:        baby1,
		Baby2:        baby2,
	}
	return info, nil
}

// GetMemberByOpenid 通过openid获取会员信息
func (c *CardService) GetMemberByOpenid(openid string) (*CardMemberInfo, error) {
	token, err := c.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(urlGetUserCards, token)
	param := map[string]interface{}{
		"card_id": c.wechat.MemberCardID,
		"openid":  openid,
	}
	req, err := c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return nil, err
	}
	result := &struct {
		Errcode  int    `json:"errcode"`
		Errmsg   string `json:"errmsg"`
		CardList []struct {
			CardID string `json:"card_id"`
			Code   string `json:"code"`
		} `json:"card_list"`
		HasShareCard bool `json:"has_share_card"`
	}{}
	_, err = c.wechat.Do(nil, req, result)
	if err != nil {
		return nil, err
	}
	if result.Errcode != 0 {
		return nil, errors.New("获取卡列表失败")
	}
	if len(result.CardList) == 0 {
		return nil, errors.New("用户还没有领取过会员卡")
	}
	cardCode := ""
	for _, card := range result.CardList {
		if strings.EqualFold(c.wechat.MemberCardID, card.CardID) {
			cardCode = card.Code
			break
		}
	}
	if len(cardCode) == 0 {
		return nil, errors.New("用户不是会员卡成员")
	}
	return c.GetMemberByCode(cardCode)
}

// unmarshalCardMemberFields 解析卡会员的参数
func unmarshalCardMemberFields(fields []CardMemberField) (mobile, gender, realName string, birthday time.Time) {
	mobile, gender, realName, birthday = "", "", genderOther, time.Now()

	for _, field := range fields {
		switch field.Name {
		case cardActivateMobile:
			mobile = field.Value
		case cardActivateBirthday:
			if strings.Contains(field.Value, "0001-01-01") {
				birthday = time.Now()
			} else if b, e := time.Parse("2006-01-02", field.Value); e == nil {
				birthday = b
			} else if b, e := time.Parse("2006-1-02", field.Value); e == nil {
				birthday = b
			} else if b, e := time.Parse("2006-01-2", field.Value); e == nil {
				birthday = b
			} else if b, e := time.Parse("2006-1-2", field.Value); e == nil {
				birthday = b
			} else {
				log.Printf("特殊的生日格式：%s", field.Value)
			}
		case cardActivateSex:
			if strings.EqualFold(field.Value, "MALE") {
				gender = genderMale
			} else if strings.EqualFold(field.Value, "FEMAIL") {
				gender = genderFemale
			}
		case cardActivateName:
			realName = field.Value
		}
	}
	return
}

func unmarshalCardMemberCustomFields(fields []CardMemberField) (baby1, baby2 string) {
	baby1, baby2 = "2000", "2000"

	for _, field := range fields {
		switch field.Name {
		case "大宝生日":
			baby1 = field.Value
		case "二宝生日":
			baby2 = field.Value
		}
	}
	return
}
