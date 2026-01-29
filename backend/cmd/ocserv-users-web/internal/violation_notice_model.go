package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	// globalVpnAccessNoticeConfig 全局 VPN 访问告警配置，受 globalVpnAccessNoticeMu 保护
	globalVpnAccessNoticeConfig *VpnAccessNoticeConfig
	// globalVpnAccessNoticeMu 保护全局 VPN 访问告警配置的并发访问
	globalVpnAccessNoticeMu sync.RWMutex
)

type VpnAccessNoticeConfig struct {
	Enabled               *bool   `json:"enabled" yaml:"enabled"` // 是否启用 VPN 访问告警, 默认为 false
	WebhookReceiver       string `json:"webhook_receiver" yaml:"webhook_receiver"`
	OnlyNotifyOnViolation *bool   `json:"only_notify_on_violation" yaml:"only_notify_on_violation"` // 是否仅在违规时发送告警，默认为 true
}

func (ac *VpnAccessNoticeConfig) setDefaults() {
	if ac.Enabled == nil {
		defaultEnabled := false
		ac.Enabled = &defaultEnabled
	}
	if ac.OnlyNotifyOnViolation == nil {
		defaultOnlyNotify := true
		ac.OnlyNotifyOnViolation = &defaultOnlyNotify
	}
}

// VpnAccessViolationPayload 是发送到 webhook 的 VPN 访问违规告警数据
type VpnAccessViolationPayload struct {
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	RemoteIP  string `json:"remote_ip"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
}

// InitializeVpnAccessNoticeConfig 初始化全局 VPN 访问告警配置
func InitializeVpnAccessNoticeConfig(config VpnAccessNoticeConfig) {
	globalVpnAccessNoticeMu.Lock()
	defer globalVpnAccessNoticeMu.Unlock()
	globalVpnAccessNoticeConfig = &config
	log.Printf("[vpn-access-notice] 初始化告警配置: enabled=%v, webhook_receiver=%s, only_notify_on_violation=%v", config.Enabled, config.WebhookReceiver, config.OnlyNotifyOnViolation)
}

// SendVpnAccessNotice 通过 webhook 发送 VPN 访问告警
// isAllowSession: 是否为允许的会话，用于判断是否需要发送
func SendVpnAccessNotice(username, remoteIP, reason string, action *VpnActionType, isAllowSession bool) error {
	globalVpnAccessNoticeMu.RLock()
	defer globalVpnAccessNoticeMu.RUnlock()

	// 检查是否启用了告警
	if globalVpnAccessNoticeConfig == nil || !*globalVpnAccessNoticeConfig.Enabled {
		return nil
	}

	// 如果配置为只在违规时发送，且当前是允许的会话，则跳过
	if globalVpnAccessNoticeConfig.OnlyNotifyOnViolation != nil && *globalVpnAccessNoticeConfig.OnlyNotifyOnViolation && isAllowSession {
		return nil
	}

	webhookURL := globalVpnAccessNoticeConfig.WebhookReceiver
	if webhookURL == "" {
		log.Println("[vpn-access-notice] webhook receiver url is empty, skip sending")
		return nil
	}

	// 确定标题
	title := fmt.Sprintf("[%s]登录VPN", username)
	// 构建告警数据
	payload := VpnAccessViolationPayload{
		Title:     title,
		Timestamp: time.Now().Format(time.RFC3339),
		Username:  username,
		RemoteIP:  remoteIP,
		Reason:    reason,
	}

	// 如果是违规会话，更新标题和 payload
	if !isAllowSession {
		title = fmt.Sprintf("[%s]违规登录VPN, 已 %s", username, string(*action))
		payload.Action = string(*action)
	}

	// 异步发送，避免阻塞主逻辑
	go func() {
		if err := sendVpnAccessWebhookAsync(webhookURL, payload); err != nil {
			log.Printf("[vpn-access-notice] 发送 webhook 告警失败: %v", err)
		}
	}()

	return nil
}

// sendVpnAccessWebhookAsync 异步发送 VPN 访问违规告警的 webhook 请求
func sendVpnAccessWebhookAsync(webhookURL string, payload VpnAccessViolationPayload) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status code %d", resp.StatusCode)
	}

	log.Printf("[vpn-access-notice] webhook 告警已发送: title=%s, user=%s, remote_ip=%s, action=%s", payload.Title, payload.Username, payload.RemoteIP, payload.Action)
	return nil
}
