package wechat

import (
	"fmt"
	"log"

	"github.com/gotit/errors"
)

/*
设置微信会员卡为跳转型一键激活时，需要按下面流程调用 1、2两个步骤是预先设置，3、4两个步骤是每次用户激活会员卡触发调用
	1、在创建/更新接口填入跳转型一键激活相关字段
	2、设置开卡字段接口（必填、选填、自定义参数）
	3、获取用户提交资料（单独接口获取用户提交资料，调用接口解码用户卡号code）
	4、调用接口激活会员卡（更新本地数据库后，通知微信）
*/

// SetActivateJump 设置微信激活后跳转连接
func (c *CardService) SetActivateJump(submitURL, levelURL, couponURL string) error {
	type CustomField struct {
		NameType string `json:"name_type"`
		URL      string `json:"url"`
	}
	activate := struct {
		AutoActivate             bool        `json:"auto_activate" form:"auto_activate"` // 是否自动激活
		WxActivate               bool        `json:"wx_activate" form:"wx_activate"`     // 微信激活
		WxActivateAfterSubmit    bool        `json:"wx_activate_after_submit"`           // 是否设置跳转型一键激活
		WxActivateAfterSubmitURL string      `json:"wx_activate_after_submit_url"`       // 用户提交信息后跳转的网页
		CustomField2             CustomField `json:"custom_field2"`
		CustomField3             CustomField `json:"custom_field3"`
	}{
		AutoActivate:             false,
		WxActivate:               true,
		WxActivateAfterSubmit:    true,
		WxActivateAfterSubmitURL: submitURL,
		CustomField2: CustomField{
			NameType: "FIELD_NAME_TYPE_LEVEL",
			URL:      levelURL,
		},
		CustomField3: CustomField{
			NameType: "FIELD_NAME_TYPE_COUPON",
			URL:      couponURL,
		},
	}
	param := map[string]interface{}{
		"card_id":     c.wechat.MemberCardID,
		"member_card": activate,
	}
	token, err := c.wechat.GetAccessToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(urlUpdateCard, token)
	req, err := c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return err
	}
	result := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	_, err = c.wechat.Do(nil, req, result)
	if err != nil {
		return err
	}
	if result.Errcode != 0 {
		return errors.Errorf("更新会员卡失败 %s", result.Errmsg)
	}
	return c.SetActivateFlag()
}

/*
SetActivateFlag 设置微信一键激活参数
	当前需要必填 姓名、性别、电话、生日、小孩数量

	type activateCardURL struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	type activateCard struct {
		CardID           string          `json:"card_id"`
		ServiceStatement activateCardURL `json:"service_statement"` // 服务声明，用于放置商户会员卡守则
		BindOldCard      activateCardURL `json:"bind_old_card"`     // 绑定老会员链接
		RequiredForm     fieldForm       `json:"required_form"`     // 会员卡激活时的必填选项
		OptionalForm     fieldForm       `json:"optional_form"`     // 会员卡激活时的选填项
	}*/
func (c *CardService) SetActivateFlag() error {
	type fieldFormSelect struct {
		Type   string   `json:"type"`   // 富文本类型 FORM_FIELD_RADIO 自定义单选, FORM_FIELD_SELECT 自定义选择项, FORM_FIELD_CHECK_BOX 自定义多选
		Name   string   `json:"name"`   // 字段名
		Values []string `json:"values"` // 选择项
	}
	type fieldForm struct {
		CanModify         bool              `json:"can_modify"`           // 当前结构（required_form或者optional_form ）内的字段是否允许用户激活后再次修改，商户设置为true时，需要接收相应事件通知处理修改事件
		RichFieldList     []fieldFormSelect `json:"rich_field_list"`      // 自定义富文本类型，包含以下三个字段，开发者可以分别在必填和选填中至多定义五个自定义选项
		CommonFieldIDList []string          `json:"common_field_id_list"` // 微信格式化的选项类型。见以下列表
	}
	required := fieldForm{
		CanModify:         false,
		RichFieldList:     []fieldFormSelect{},
		CommonFieldIDList: []string{cardActivateMobile, cardActivateSex, cardActivateName, cardActivateBirthday},
	}
	optional := fieldForm{
		CanModify: false,
		RichFieldList: []fieldFormSelect{
			{Type: "FORM_FIELD_RADIO", Name: "大宝年龄", Values: []string{"0~1岁", "1~3岁", "3~6岁", "6岁以上"}},
			{Type: "FORM_FIELD_RADIO", Name: "二宝年龄", Values: []string{"0~1岁", "1~3岁", "3~6岁", "6岁以上"}},
		},
		CommonFieldIDList: []string{},
	}

	activateFlag := map[string]interface{}{
		"card_id":       c.wechat.MemberCardID,
		"required_form": required,
		"optional_form": optional,
	}
	token, err := c.wechat.GetAccessToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(urlActivateCardFlag, token)
	req, err := c.wechat.NewRequest("POST", url, activateFlag)
	if err != nil {
		return err
	}
	result := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	_, err = c.wechat.Do(nil, req, result)
	if err != nil {
		return err
	}
	if result.Errcode != 0 {
		return errors.Errorf("设置微信一键激活参数失败 %s", result.Errmsg)
	}
	return nil
}

// GetUseSubmitParam 获取微信指定用户提交的激活信息
func (c *CardService) GetUseSubmitParam(encryptCode, openid, activateTicket string) (*CardMemberInfo, error) {
	// 用解码结果获取激活信息
	param := map[string]interface{}{
		"activate_ticket": activateTicket,
	}
	token, err := c.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(urlGetCardUserActivate, token)
	req, err := c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return nil, err
	}
	memberInfoResult := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		Info    struct {
			CommonFieldList []CardMemberField `json:"common_field_list"`
			CustomFieldList []CardMemberField `json:"custom_field_list"`
		} `json:"info"`
	}{}
	_, err = c.wechat.Do(nil, req, memberInfoResult)
	if err != nil {
		return nil, err
	}
	if memberInfoResult.Errcode != 0 {
		log.Printf("获取激活信息错误了，这个错误八成是微信是bug %v", memberInfoResult)
		return nil, nil
	}

	// 调用微信接口解码卡号
	param = map[string]interface{}{
		"encrypt_code": encryptCode,
	}
	token, err = c.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	url = fmt.Sprintf(urlCardCodeDecrypt, token)
	req, err = c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return nil, err
	}
	cardCodeResult := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		Code    string `json:"code"`
	}{}
	_, err = c.wechat.Do(nil, req, cardCodeResult)
	if err != nil {
		return nil, err
	}
	if cardCodeResult.Errcode != 0 {
		return nil, errors.Errorf("解码会员卡号出错 %s", cardCodeResult.Errmsg)
	}

	// 调用微信接口激活会员卡
	param = map[string]interface{}{
		"membership_number": cardCodeResult.Code,
		"code":              cardCodeResult.Code,
		"card_id":           c.wechat.MemberCardID,
	}
	token, err = c.wechat.GetAccessToken()
	if err != nil {
		return nil, err
	}
	url = fmt.Sprintf(urlActivateMemberCard, token)
	req, err = c.wechat.NewRequest("POST", url, param)
	if err != nil {
		return nil, err
	}
	activateResult := &struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	_, err = c.wechat.Do(nil, req, activateResult)
	if err != nil {
		return nil, err
	}
	if activateResult.Errcode != 0 {
		return nil, errors.Errorf("获取用户提交的激活信息失败 %s", activateResult.Errmsg)
	}

	userInfo, err := c.wechat.User.GetUserInfoByOpenid(openid)
	if err != nil {
		return nil, err
	}
	mobileNumber, gender, realname, birthday := unmarshalCardMemberFields(memberInfoResult.Info.CommonFieldList)
	baby1, baby2 := unmarshalCardMemberCustomFields(memberInfoResult.Info.CustomFieldList)
	memberInfo := &CardMemberInfo{
		CardID:       c.wechat.MemberCardID,
		Openid:       openid,
		Unionid:      userInfo.Unionid,
		CardCode:     cardCodeResult.Code,
		Nickname:     userInfo.Nickname,
		MobileNumber: mobileNumber,
		Gender:       gender,
		RealName:     realname,
		Birthday:     birthday,
		MemberStatus: memberCardStatusNormal,
		Baby1:        baby1,
		Baby2:        baby2,
	}
	return memberInfo, nil
}
