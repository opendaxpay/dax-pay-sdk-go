# DaxPay Open SDK for Go

DaxPay 开放支付平台 Go SDK，封装支付下单、关闭、退款、订单查询与回调验签。

> **适配 DaxPay Open ≥ 1.0** · **Go 1.21+** · LGPL-3.0 · 零第三方依赖（标准库实现）

## 功能

- RSA 双向签名（SHA256withRSA），自动签名请求 / 验签响应与回调
- 走 JSON 签名路径，与开源版后端 `reqTime` 契约对齐
- 核心支付接口：pay / close / refund / query pay-order / query refund-order
- 异步回调验签
- `context.Context` 支持

## 安装（源码引入）

```bash
go get github.com/opendaxpay/dax-pay-sdk-go@main
```

```go
import "github.com/opendaxpay/dax-pay-sdk-go/daxpay"
```

## 快速开始

```go
config := &daxpay.Config{
    ServiceUrl: "https://sandbox.daxpay.cn",
    MchNo:      "M200000001",
    AppId:      "APP001",
    PrivateKey: merchantPrivateKeyPem,   // PEM 文本
    PublicKey:  platformPublicKeyPem,    // PEM 文本
    Timeout:    30 * time.Second,
}
client := daxpay.NewClient(config)

// 支付下单
result, err := client.Pay(ctx, &daxpay.PayParam{
    BizOrderNo: "PAY20250805001",
    Title:      "测试商品",
    Amount:     100,            // 分
    Method:     "wechat_qr",
    NotifyUrl:  "https://example.com/notify",
})

// 回调验签
// ok := client.VerifyNotice(rawBody)
```

> 完整可运行示例见 [`examples/pay/main.go`](examples/pay/main.go)。

## 接口文档

- [接入准备](https://doc.open.daxpay.cn/api/getting-started) · [签名规则](https://doc.open.daxpay.cn/api/signature)
- 黄金测试向量：见 [`daxpay/golden_test.go`](daxpay/golden_test.go)（与后端签名契约同源断言）

## License

LGPL-3.0，与主仓库 [DaxPay Open](https://gitee.com/dromara/dax-pay) 同协议。
