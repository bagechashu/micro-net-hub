package web

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ocserv-users/internal"
)

// ViolationResponse represents the response structure for violation queries
type ViolationResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	RemoteIP  string `json:"remote_ip"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
}

// ViolationsListResponse represents paginated violations response
type ViolationsListResponse struct {
	Violations []ViolationResponse `json:"violations"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	PageCount  int64               `json:"page_count"`
	Message    string              `json:"message,omitempty"`
}

// StatsCache is a cache for stats queries with TTL.
type StatsCache struct {
	data      map[string]interface{}
	timestamp time.Time
	mu        sync.RWMutex
}

var (
	statsCache = &StatsCache{}
	statsTTL   = 30 * time.Second
)

// GetViolationsHandler handles API requests to retrieve violations with pagination
func GetViolationsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate query parameters
	username := r.URL.Query().Get("username")
	action := r.URL.Query().Get("action")
	daysStr := r.URL.Query().Get("days")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := parseLimit(limitStr)
	offset := parseOffset(offsetStr)

	startTime, endTime, useCustomRange, err := parseCustomDateRange(startDateStr, endDateStr)
	if err != nil {
		log.Printf("[api] invalid date format: startDate=%s, endDate=%s", startDateStr, endDateStr)
		sendError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	timeBack := parseDays(daysStr)
	query := internal.ViolationQuery{
		TimeBack:       timeBack,
		StartTime:      startTime,
		EndTime:        endTime,
		UseCustomRange: useCustomRange,
		Username:       username,
		Action:         action,
		Limit:          limit,
		Offset:         offset,
	}

	violations, total, err := internal.GetViolations(query)
	if err != nil {
		log.Printf("[api] failed to get violations: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to retrieve violations")
		return
	}

	// Convert to response format
	responses := make([]ViolationResponse, len(violations))
	for i, v := range violations {
		responses[i] = ViolationResponse{
			ID:        v.ID,
			Username:  v.Username,
			RemoteIP:  v.RemoteIP,
			Action:    v.Action,
			Reason:    v.Reason,
			Timestamp: v.Timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	pageCount := (total + int64(limit) - 1) / int64(limit)
	listResp := ViolationsListResponse{
		Violations: responses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		PageCount:  pageCount,
	}
	sendSuccess(w, "success", listResp)
}

// GetViolationStatsHandler returns action-based violation statistics
// Optionally filters by username with caching for improved performance
func GetViolationStatsHandler(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	username := r.URL.Query().Get("username")

	// Check cache only if no username filter and no custom range (user-specific stats are more volatile)
	if username == "" && startDateStr == "" && endDateStr == "" {
		statsCache.mu.RLock()
		if time.Since(statsCache.timestamp) < statsTTL && statsCache.data != nil {
			defer statsCache.mu.RUnlock()
			sendSuccess(w, "success", statsCache.data)
			return
		}
		statsCache.mu.RUnlock()
	}

	startTime, endTime, useCustomRange, err := parseCustomDateRange(startDateStr, endDateStr)
	if err != nil {
		log.Printf("[api] invalid date format: startDate=%s, endDate=%s", startDateStr, endDateStr)
		sendError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	timeBack := parseDays(daysStr)
	query := internal.ViolationQuery{
		StartTime:      startTime,
		EndTime:        endTime,
		TimeBack:       timeBack,
		UseCustomRange: useCustomRange,
		Username:       username,
	}

	stats, err := internal.GetViolationStats(query)
	if err != nil {
		log.Printf("[api] failed to get violation stats: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to retrieve statistics")
		return
	}

	// Update cache if no username filter and no custom range
	if username == "" && !useCustomRange {
		statsCache.mu.Lock()
		statsCache.data = stats
		statsCache.timestamp = time.Now()
		statsCache.mu.Unlock()
	}

	sendSuccess(w, "success", stats)
}

// GetViolationUsersHandler returns top users with their action breakdown
func GetViolationUsersHandler(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	timeBack := parseDays(daysStr)

	stats, err := internal.GetViolationUsers(timeBack)
	if err != nil {
		log.Printf("[api] failed to get user stats: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to retrieve user statistics")
		return
	}
	sendSuccess(w, "success", stats)
}

func ClearViolationOldDataHandler(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	timeBack := parseDays(daysStr)

	err := internal.ClearOldViolations(timeBack)
	if err != nil {
		log.Printf("[api] violation cleared old data failed: %v", err)
		sendError(w, http.StatusInternalServerError, "violation cleared old data failed")
		return
	}
	sendSuccess(w, "violation cleared old data successfully", nil)
}

// parseCustomDateRange Handle custom date range if provided
func parseCustomDateRange(startDateStr, endDateStr string) (startTime time.Time, endTime time.Time, enable bool, err error) {
	enable = false
	if startDateStr == "" && endDateStr == "" {
		return time.Time{}, time.Time{}, enable, nil
	}
	startTime, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, enable, err
	}
	endTime, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, enable, err
	}
	// Add 24 hours to end date to include the whole day
	endTime = endTime.Add(24 * time.Hour)
	enable = true
	return startTime, endTime, enable, nil
}

// parseDays parses the days query parameter and returns a time.Duration
func parseDays(daysStr string) time.Duration {
	if daysStr == "" {
		return time.Duration(0)
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		return time.Duration(0)
	}

	return time.Duration(days*24) * time.Hour
}

func parseLimit(limitStr string) int {
	if limitStr == "" {
		return 20
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 20
	}

	if limit > 1000 {
		limit = 1000
	}

	return limit
}

func parseOffset(offsetStr string) int {
	if offsetStr == "" {
		return 0
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0
	}

	return offset
}
