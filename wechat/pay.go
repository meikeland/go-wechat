package wechat

// PayService 订单支付服务
type PayService service

const (
	// 支付URL

	urlJsapiTicket = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi" // 获取jsapiTicket url
	urlUnifyOrder  = "https://api.mch.weixin.qq.com/pay/unifiedorder"                                // 微信统一下单接口
	urlOrderQuery  = "https://api.mch.weixin.qq.com/pay/orderquery"                                  // 微信订单查询接口

	// 支付类型

	tradeTypeJsapi    = "JSAPI"    // 公众号支付
	tradeTypeNative   = "NATIVE"   // 原生扫码支付
	tradeTypeApp      = "APP"      // app支付
	tradeTypeMicropay = "MICROPAY" // 刷卡支付，刷卡支付有单独的支付接口，不调用统一下单接口

	// 支付结果

	tradeStateSUCCESS    = "SUCCESS"    // 支付成功
	tradeStateREFUND     = "REFUND"     // 转入退款
	tradeStateNOTPAY     = "NOTPAY"     // 未支付
	tradeStateCLOSED     = "CLOSED"     // 已关闭
	tradeStateREVOKED    = "REVOKED"    // 已撤销（刷卡支付）
	tradeStateUSERPAYING = "USERPAYING" // 用户支付中
	tradeStatePAYERROR   = "PAYERROR"   // 支付失败(其他原因，如银行返回失败)
)
