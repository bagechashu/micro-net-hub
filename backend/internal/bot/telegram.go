package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// telegramType Telegram 的类型标识，同时作为 config 中 token 的归属类型
const telegramType = "telegram"

// telegramSpec Telegram 实例的配置规格
var telegramSpec = TypeSpec{
	Type:          telegramType,
	Label:         "Telegram",
	RequiredKeys:  []string{"token"},
	SensitiveKeys: []string{"token"},
}

// updateTimeoutSeconds 长轮询等待时长（秒）
const updateTimeoutSeconds = 60

// TelegramProvider 基于 Telegram Bot API 的 BotProvider 实现
type TelegramProvider struct {
	id      string
	name    string
	bot     *tgbotapi.BotAPI
	handler func(Message)
}

var _ BotProvider = (*TelegramProvider)(nil)

// NewTelegramProvider 用 Bot Token 创建 Telegram provider。
//
// 返回具体类型而非接口（Go 惯例：返回具体类型、接受接口），注册进 Registry 时由
// newTelegramProvider 适配成 Factory 签名。
func NewTelegramProvider(id, name, token string) (*TelegramProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram 实例 [%s] 的 token 为空", id)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram 实例 [%s] 认证失败: %w", id, err)
	}
	bot.Debug = false
	log.Printf("[Telegram:%s] 已认证账号 %s", id, bot.Self.UserName)

	// 未配置显示名时回落到 Bot 自身的用户名，保证日志与面板里有可读标识
	displayName := name
	if displayName == "" {
		displayName = bot.Self.UserName
	}

	return &TelegramProvider{id: id, name: displayName, bot: bot}, nil
}

// newTelegramProvider 把 NewTelegramProvider 适配为 Registry 所需的 Factory。
//
// token 的存在性由 Registry 依据 telegramSpec.RequiredKeys 前置校验，此处直接取用。
func newTelegramProvider(id, name string, config map[string]string) (BotProvider, error) {
	return NewTelegramProvider(id, name, config["token"])
}

// ID 返回实例唯一标识
func (t *TelegramProvider) ID() string { return t.id }

// Type 返回 Bot 类型
func (t *TelegramProvider) Type() string { return telegramType }

// Name 返回显示名称
func (t *TelegramProvider) Name() string { return t.name }

// RegisterHandler 注册消息处理器
func (t *TelegramProvider) RegisterHandler(handler func(Message)) {
	t.handler = handler
}

// Start 启动长轮询监听，阻塞直到 ctx 被取消。
//
// 退出前必须 StopReceivingUpdates：tgbotapi 内部为 GetUpdatesChan 起了独立
// goroutine，只取消 ctx 不会让它退出，热加载反复重建实例会持续泄漏。
func (t *TelegramProvider) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = updateTimeoutSeconds

	updates := t.bot.GetUpdatesChan(u)
	defer t.bot.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("telegram 实例 [%s] 的 update 通道已关闭", t.id)
			}
			t.processUpdate(update)
		}
	}
}

// processUpdate 将 Telegram update 转换为统一 Message 并交给处理器
func (t *TelegramProvider) processUpdate(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	msg := Message{
		BotID:     t.id,
		BotType:   telegramType,
		UserID:    strconv.FormatInt(update.Message.From.ID, 10),
		Username:  displayUsername(update.Message.From),
		ChannelID: strconv.FormatInt(update.Message.Chat.ID, 10),
		ChatType:  update.Message.Chat.Type, // private / group / supergroup：审批等安全边界据此判定
		Content:   update.Message.Text,
	}

	if t.handler != nil {
		t.handler(msg)
	}
}

// displayUsername 构建可读用户名：优先 UserName，其次 FirstName + LastName，最后用数字 ID
func displayUsername(from *tgbotapi.User) string {
	if from.UserName != "" {
		return from.UserName
	}

	nameParts := []string{from.FirstName}
	if from.LastName != "" {
		nameParts = append(nameParts, from.LastName)
	}
	if name := strings.TrimSpace(strings.Join(nameParts, " ")); name != "" {
		return name
	}
	return strconv.FormatInt(from.ID, 10)
}

// SendMessage 发送普通文本消息
func (t *TelegramProvider) SendMessage(ctx context.Context, toID string, text string) error {
	return t.send(toID, text, "")
}

// SendHTMLMessage 发送 HTML 格式消息
func (t *TelegramProvider) SendHTMLMessage(ctx context.Context, toID string, html string) error {
	return t.send(toID, html, tgbotapi.ModeHTML)
}

// SendTyping 向目标会话发送"正在输入"状态。
//
// Telegram 的 ChatAction 是一次性状态，约 5 秒后自动失效，长耗时请求需要由调用方
// 按 TypingInterval 反复调用。ctx 当前仅用于接口契约一致（tgbotapi 的发送不接收
// context），后续若换用支持 ctx 的客户端会用到。
func (t *TelegramProvider) SendTyping(_ context.Context, toID string) error {
	chatID, err := strconv.ParseInt(toID, 10, 64)
	if err != nil {
		return fmt.Errorf("非法的 telegram chat id %q: %w", toID, err)
	}

	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	// sendChatAction 的响应体是 {"ok":true,"result":true}，result 是 bool 而非 Message，
	// 因此不能走 bot.Send（它会强转 result 为 Message，报 "cannot unmarshal bool into
	// tgbotapi.Message"），改用 bot.Request 只校验 ok、不解析 result。
	if _, err := t.bot.Request(action); err != nil {
		return fmt.Errorf("telegram 实例 [%s] 发送输入状态失败: %w", t.id, err)
	}
	return nil
}

// send 按指定 ParseMode 发送消息到目标会话
func (t *TelegramProvider) send(toID, text, parseMode string) error {
	chatID, err := strconv.ParseInt(toID, 10, 64)
	if err != nil {
		return fmt.Errorf("非法的 telegram chat id %q: %w", toID, err)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	if _, err := t.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram 实例 [%s] 发送消息失败: %w", t.id, err)
	}
	return nil
}
