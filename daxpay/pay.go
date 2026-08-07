package daxpay

import (
	"context"
	"encoding/json"
	"fmt"
)

// TerminalInfo 终端信息（线下场景）— 对照契约 6.6
type TerminalInfo struct {
	TerminalNo string  `json:"terminalNo,omitempty"`
	StoreNo    string  `json:"storeNo,omitempty"`
	OperatorId string  `json:"operatorId,omitempty"`
	DeviceName string  `json:"deviceName,omitempty"`
	DeviceIp   string  `json:"deviceIp,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
}

// GoodsDetail 商品明细 — 对照契约 6.6
type GoodsDetail struct {
	GoodsId     string `json:"goodsId"`
	GoodsName   string `json:"goodsName"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unitPrice"` // 分
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	ShowUrl     string `json:"showUrl,omitempty"`
}

// PayParam 支付下单请求参数（公共字段由 Execute 自动注入）— 对照契约 6.1
type PayParam struct {
	BizOrderNo   string        `json:"bizOrderNo"`            // 商户订单号（必填）
	Title        string        `json:"title"`                 // 支付标题（必填）
	Description  string        `json:"description,omitempty"`
	Amount       int64         `json:"amount"`                // 支付金额，分（必填）
	Currency     string        `json:"currency,omitempty"`
	Product      string        `json:"product,omitempty"`
	Method       string        `json:"method,omitempty"`
	Capability   string        `json:"capability,omitempty"`
	OpenId       string        `json:"openId,omitempty"`
	ChannelAppId string        `json:"channelAppId,omitempty"`
	AuthCode     string        `json:"authCode,omitempty"`
	LimitPay     []string      `json:"limitPay,omitempty"`
	ExtraParam   string        `json:"extraParam,omitempty"`
	GoodsDetail  []GoodsDetail `json:"goodsDetail,omitempty"`
	NotifyUrl    string        `json:"notifyUrl,omitempty"`
	ReturnUrl    string        `json:"returnUrl,omitempty"`
	Attach       string        `json:"attach,omitempty"`
	ExpiredTime  string        `json:"expiredTime,omitempty"` // GMT+8 yyyy-MM-dd HH:mm:ss
	Terminal     *TerminalInfo `json:"terminal,omitempty"`
}

// PayResult 支付下单响应结果
type PayResult struct {
	OrderId     int64  `json:"orderId"`
	BizOrderNo  string `json:"bizOrderNo"`
	OrderNo     string `json:"orderNo"`
	TradeNo     string `json:"tradeNo"`
	Status      string `json:"status"`      // wait/progress/success/close/cancel/fail/timeout
	PayBody     string `json:"payBody"`     // 二维码链接/调起参数/跳转 URL
	PayBodyType string `json:"payBodyType"` // code_url/pay_info/redirect_url
}

// Pay 支付下单便捷方法 — POST /unipay/pay
func (c *Client) Pay(ctx context.Context, param *PayParam) (*PayResult, error) {
	raw, err := c.Execute(ctx, "/unipay/pay", param)
	if err != nil {
		return nil, err
	}
	var pr PayResult
	if err := json.Unmarshal(raw.Data, &pr); err != nil {
		return nil, fmt.Errorf("支付结果解析失败: %w", err)
	}
	return &pr, nil
}
