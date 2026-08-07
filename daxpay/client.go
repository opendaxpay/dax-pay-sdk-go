package daxpay

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client DaxPay SDK 客户端 — 对照 sdk-contract.md 第十节
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient 创建客户端
func NewClient(config Config) *Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Client{
		config:     config,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Execute 通用执行入口：自动填充公共参数 → JSON 签名 → POST → 验签 → 返回 RawResult
// 走 JSON 签名路径（reqTime 已序列化为 GMT+8 字面量），与后端验签一致
// param 可传入 struct（如 *PayParam）或 map，内部统一转 map 注入公共字段
func (c *Client) Execute(ctx context.Context, path string, param interface{}) (*RawResult, error) {
	// 统一转 map（struct 或 map 均可；金额单位为分，float64 中间表示对分金额无精度损失）
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %w", err)
	}
	var paramMap map[string]interface{}
	if err := json.Unmarshal(paramBytes, &paramMap); err != nil {
		return nil, fmt.Errorf("请求解析失败: %w", err)
	}
	// 注入公共字段
	if _, ok := paramMap["mchNo"]; !ok {
		paramMap["mchNo"] = c.config.MchNo
	}
	if _, ok := paramMap["appId"]; !ok && c.config.AppId != "" {
		paramMap["appId"] = c.config.AppId
	}
	if _, ok := paramMap["reqId"]; !ok {
		paramMap["reqId"] = newUUID()
	}
	if _, ok := paramMap["reqTime"]; !ok {
		paramMap["reqTime"] = nowGmt8()
	}
	if _, ok := paramMap["nonceStr"]; !ok {
		paramMap["nonceStr"] = newNonce()
	}

	// 走 JSON 签名路径
	jsonForSign, err := json.Marshal(paramMap)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %w", err)
	}
	signStr, err := BuildSignStr(string(jsonForSign))
	if err != nil {
		return nil, fmt.Errorf("签名串构造失败: %w", err)
	}
	sign, err := RsaSign(signStr, c.config.PrivateKey)
	if err != nil {
		return nil, err
	}
	paramMap["sign"] = sign
	body, err := json.Marshal(paramMap)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %w", err)
	}

	url := strings.TrimRight(c.config.ServiceUrl, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result RawResult
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %w", err)
	}
	// 用原始 body 字符串验签
	if result.Sign != "" {
		verifyStr, err := BuildSignStr(string(rawBody))
		if err != nil {
			return nil, fmt.Errorf("验签串构造失败: %w", err)
		}
		ok, err := RsaVerify(verifyStr, result.Sign, c.config.PublicKey)
		if err != nil || !ok {
			return nil, errors.New("响应验签失败")
		}
	}
	if result.Code != 0 {
		return nil, &BizError{Code: result.Code, Msg: result.Msg}
	}
	return &result, nil
}

// VerifyNotice 回调验签（原始 HTTP body 字符串）— 对照契约第八节
func (c *Client) VerifyNotice(rawBody string) bool {
	var probe struct {
		Sign string `json:"sign"`
	}
	if json.Unmarshal([]byte(rawBody), &probe) != nil {
		return false
	}
	if probe.Sign == "" {
		return false
	}
	verifyStr, err := BuildSignStr(rawBody)
	if err != nil {
		return false
	}
	ok, _ := RsaVerify(verifyStr, probe.Sign, c.config.PublicKey)
	return ok
}

// nowGmt8 当前时间的 GMT+8 字面量（yyyy-MM-dd HH:mm:ss）— 对照后端 @JsonFormat(GMT+8)
func nowGmt8() string {
	return time.Now().UTC().Add(8 * time.Hour).Format("2006-01-02 15:04:05")
}

// newUUID 生成 UUID v4
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// newNonce 生成 32 位随机 hex
func newNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
