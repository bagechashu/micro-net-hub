package radiusappr

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"micro-net-hub/internal/config"
)

// weekdayNames 星期缩写到 time.Weekday 的映射, 供审批时间窗口解析使用.
var weekdayNames = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

// locationCache 时区解析结果缓存, 避免每次认证都解析一次时区字符串.
var locationCache sync.Map // map[string]*time.Location

// loadLocation 解析时区, 留空或解析失败时回落到本地时区.
func loadLocation(name string) *time.Location {
	if name == "" {
		return time.Local
	}
	if loc, ok := locationCache.Load(name); ok {
		return loc.(*time.Location)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		// 时区名不合法属配置错误, 记日志后回落到本地时区, 避免认证链路直接不可用
		return time.Local
	}
	locationCache.Store(name, loc)
	return loc
}

// InApprovalWindow 判断给定时刻是否落在需要人工审批的时间窗口内.
//
// 判定规则:
//   - 审批未开启时恒为 false;
//   - 未配置任何时间窗口时视为全天都需要审批;
//   - days 为空表示每天生效; start 与 end 为 HH:MM, 当 end 不晚于 start 时视为跨天窗口;
//   - 跨天窗口的回溯由"候选取今天与昨天两个锚定日"实现, 因此 22:00-06:00 这类窗口
//     在凌晨时段也能正确命中.
func InApprovalWindow(now time.Time, approval *config.RadiusApproval) (bool, error) {
	if approval == nil || !approval.Enable {
		return false, nil
	}

	if len(approval.TimeWindows) == 0 {
		return true, nil
	}

	loc := loadLocation(approval.Timezone)
	local := now.In(loc)

	for i, window := range approval.TimeWindows {
		startOffset, start, err := parseHHMM(window.Start, fmt.Sprintf("time-windows[%d].start", i))
		if err != nil {
			return false, err
		}
		endOffset, end, err := parseHHMM(window.End, fmt.Sprintf("time-windows[%d].end", i))
		if err != nil {
			return false, err
		}

		// 锚定日: 今天与昨天, 覆盖跨天窗口落在凌晨的情况
		for _, dayOffset := range []int{0, -1} {
			anchor := local.AddDate(0, 0, dayOffset)
			matched, err := matchWindowDays(window.Days, anchor.Weekday())
			if err != nil {
				return false, fmt.Errorf("time-windows[%d]: %w", i, err)
			}
			if !matched {
				continue
			}

			startAt := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, loc).Add(startOffset)
			// end 不晚于 start 时视为跨天, 结束时刻顺延一天
			endDay := 0
			if end <= start {
				endDay = 1
			}
			endAt := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, loc).
				Add(endOffset + time.Duration(endDay)*24*time.Hour)

			if !local.Before(startAt) && local.Before(endAt) {
				return true, nil
			}
		}
	}

	return false, nil
}

// matchWindowDays 判断某个星期是否命中窗口的 days 配置; days 为空表示每天命中.
func matchWindowDays(days []string, wd time.Weekday) (bool, error) {
	if len(days) == 0 {
		return true, nil
	}
	for _, day := range days {
		key := strings.ToLower(strings.TrimSpace(day))
		if key == "" {
			continue
		}
		expected, ok := weekdayNames[key]
		if !ok {
			return false, fmt.Errorf("无法识别的星期 %q, 可选值: sun, mon, tue, wed, thu, fri, sat", day)
		}
		if expected == wd {
			return true, nil
		}
	}
	return false, nil
}

// parseHHMM 把 HH:MM 解析为当日的偏移量, 同时返回分钟数用于跨天判定.
func parseHHMM(value, field string) (time.Duration, int, error) {
	text := strings.TrimSpace(value)
	parsed, err := time.Parse("15:04", text)
	if err != nil {
		return 0, 0, fmt.Errorf("%s 时间格式非法(%q), 应为 HH:MM", field, value)
	}
	minutes := parsed.Hour()*60 + parsed.Minute()
	return time.Duration(minutes) * time.Minute, minutes, nil
}
