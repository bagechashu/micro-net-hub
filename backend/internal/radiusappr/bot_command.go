package radiusappr

import (
	"fmt"
	"strconv"
	"strings"
)

// 审批指令类型
const (
	CommandUnknown = ""        // 非审批指令
	CommandHelp    = "help"    // 帮助
	CommandID      = "id"      // 查看自身会话标识
	CommandPending = "pending" // 查看待审批列表
	CommandApprove = "approve" // 通过
	CommandReject  = "reject"  // 拒绝
	CommandGrant   = "grant"   // 应急放行
)

// 默认的 /pending 返回条数
const defaultPendingLimit = 10

// Command 解析后的审批指令
type Command struct {
	Kind       string // 指令类型, 见 CommandXxx 常量
	RequestID  uint   // 申请单 ID(/approve, /reject)
	Username   string // 目标用户名(/grant)
	TTLMinutes int    // 放行时长(分钟, /grant)
	Limit      int    // 列表条数(/pending)
	Reason     string // 审批原因(/approve, /reject)
}

// ParseCommand 解析审批指令文本.
//
// 支持格式:
//
//	/approve <申请ID> [原因]
//	/reject <申请ID> [原因]
//	/grant <用户名> [分钟数]
//	/pending [条数]
//	/id | /help | /start
//
// 非指令文本返回 Kind 为 CommandUnknown 且 err 为 nil; 指令参数非法时返回 err,
// 由调用方回执用法提示.
func ParseCommand(text string) (Command, error) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return Command{}, nil
	}

	head := strings.ToLower(fields[0])
	if !strings.HasPrefix(head, "/") {
		return Command{}, nil
	}
	// 群聊中指令通常写作 /approve@some_bot, 统一剥离 @botname 后缀
	if at := strings.Index(head, "@"); at >= 0 {
		head = head[:at]
	}
	head = strings.TrimPrefix(head, "/")

	switch head {
	case "help", "start":
		return Command{Kind: CommandHelp}, nil
	case "id":
		return Command{Kind: CommandID}, nil
	case "pending":
		cmd := Command{Kind: CommandPending, Limit: defaultPendingLimit}
		if len(fields) > 1 {
			if n, err := strconv.Atoi(fields[1]); err == nil && n > 0 {
				cmd.Limit = n
			}
		}
		return cmd, nil
	case CommandApprove, CommandReject:
		if len(fields) < 2 {
			return Command{}, fmt.Errorf("用法: /%s <申请ID> [原因]", head)
		}
		id, err := parseRequestID(fields[1])
		if err != nil {
			return Command{}, err
		}
		cmd := Command{Kind: head, RequestID: id}
		if len(fields) > 2 {
			cmd.Reason = truncate(strings.Join(fields[2:], " "), 200)
		}
		return cmd, nil
	case CommandGrant:
		if len(fields) < 2 {
			return Command{}, fmt.Errorf("用法: /grant <用户名> [分钟数]")
		}
		cmd := Command{Kind: CommandGrant, Username: truncate(fields[1], 50), TTLMinutes: 0}
		if len(fields) > 2 {
			if minutes, err := strconv.Atoi(fields[2]); err == nil && minutes > 0 {
				cmd.TTLMinutes = minutes
			}
		}
		return cmd, nil
	default:
		return Command{}, nil
	}
}

// parseRequestID 解析申请单 ID, 兼容 #12 这样的写法
func parseRequestID(raw string) (uint, error) {
	text := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	id, err := strconv.ParseUint(text, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("申请ID 非法: %q", raw)
	}
	return uint(id), nil
}

// truncate 按字符数截断文本, 保护数据库字段长度
func truncate(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max])
}
