package wechat

// CardService 处理与卡券相关的API，包括会员卡和优惠券
type CardService service

const (
	// 微信卡券相关URL

	urlGetCardList         = "https://api.weixin.qq.com/card/batchget?access_token=%s"                        // 获取帐号卡列表
	urlGetCardInfo         = "https://api.weixin.qq.com/card/get?access_token=%s"                             // 获取卡信息
	urlGetUserCards        = "https://api.weixin.qq.com/card/user/getcardlist?access_token=%s"                // 获取用户的卡券
	urlGetCardUserInfo     = "https://api.weixin.qq.com/card/membercard/userinfo/get?access_token=%s"         // 获取卡券用户填写的信息
	urlGetCardUserActivate = "https://api.weixin.qq.com/card/membercard/activatetempinfo/get?access_token=%s" // 获取微信会员激活会员卡的输入字段参数

	urlUpdateCard         = "https://api.weixin.qq.com/card/update?access_token=%s"                          // 更新卡信息
	urlActivateMemberCard = "https://api.weixin.qq.com/card/membercard/activate?access_token=%s"             // 激活用户指定的会员卡接口
	urlActivateCardFlag   = "https://api.weixin.qq.com/card/membercard/activateuserform/set?access_token=%s" // 设置会员卡激活输入字段接口
	urlCardCodeDecrypt    = "https://api.weixin.qq.com/card/code/decrypt?access_token=%s"                    // 微信卡券卡号解码接口

	// 微信卡券类型

	cardDiscount   = "DISCOUNT"    // 卡券类型 优惠券
	cardMemberCard = "MEMBER_CARD" // 卡券类型 会员卡

	// 激活会员卡可选字段

	cardActivateMobile           = "USER_FORM_INFO_FLAG_MOBILE"            // 手机号
	cardActivateSex              = "USER_FORM_INFO_FLAG_SEX"               // 性别
	cardActivateName             = "USER_FORM_INFO_FLAG_NAME"              // 姓名
	cardActivateBirthday         = "USER_FORM_INFO_FLAG_BIRTHDAY"          // 生日
	cardActivateIDcard           = "USER_FORM_INFO_FLAG_IDCARD"            // 身份证
	cardActivateEmail            = "USER_FORM_INFO_FLAG_EMAIL"             // 邮箱
	cardActivateLocation         = "USER_FORM_INFO_FLAG_LOCATION"          // 详细地址
	cardActivateEducationBackgro = "USER_FORM_INFO_FLAG_EDUCATION_BACKGRO" // 教育背景
	cardActivateIndustry         = "USER_FORM_INFO_FLAG_INDUSTRY"          // 行业
	cardActivateIncome           = "USER_FORM_INFO_FLAG_INCOME"            // 收入
	cardActivateHabit            = "USER_FORM_INFO_FLAG_HABIT"             // 兴趣爱好

	// 性别

	genderFemale = "FEMALE"  // 女
	genderMale   = "MALE"    // 男
	genderOther  = "UNKNOWN" // 未知

	// 用户的会员卡状态

	memberCardStatusNormal      = "NORMAL"       // 正常
	memberCardStatusExpire      = "EXPIRE"       // 已过期
	memberCardStatusGifting     = "GIFTING"      // 转赠中
	memberCardStatusGiftSucc    = "GIFT_SUCC"    // 转赠成功
	memberCardStatusGiftTimeout = "GIFT_TIMEOUT" // 转赠超时
	memberCardStatusDelete      = "DELETE"       // 已删除
	memberCardStatusUnavailable = "UNAVAILABLE"  // 已失效

	// 微信卡券状态

	cardStatusNotVerify  = "CARD_STATUS_NOT_VERIFY"  // 待审核
	cardStatusVerifyFail = "CARD_STATUS_VERIFY_FAIL" // 审核失败
	cardStatusVerifyOk   = "CARD_STATUS_VERIFY_OK"   // 通过审核
	cardStatusDelete     = "CARD_STATUS_DELETE"      // 卡券被商户删除
	cardStatusDispatch   = "CARD_STATUS_DISPATCH"    // 在公众平台投放过的卡券
)
