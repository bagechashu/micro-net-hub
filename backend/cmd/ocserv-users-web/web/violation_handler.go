package web

import (
	"encoding/json"
	"log"
	"net/http"
	"ocserv-users/internal"
	"strconv"
	"time"
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
	Data       []ViolationResponse `json:"data"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	PageCount  int64               `json:"page_count"`
	Message    string              `json:"message,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// GetViolationsHandler handles API requests to retrieve violations with pagination
func GetViolationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse and validate query parameters
	username := r.URL.Query().Get("username")
	action := r.URL.Query().Get("action")

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d >= 0 {
			if d > 365 { // Max 1 year
				d = 365
			}
			if d == 0 { // 0 means all
				days = 0
			} else {
				days = d
			}
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 1000 {
				l = 1000
			}
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	timeBack := time.Duration(0)
	if days > 0 {
		timeBack = time.Duration(days*24) * time.Hour
	}

	query := internal.ViolationQuery{
		TimeBack: timeBack,
		Username: username,
		Action:   action,
		Limit:    limit,
		Offset:   offset,
	}

	violations, total, err := internal.GetViolations(query)
	if err != nil {
		log.Printf("[api] failed to get violations: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve violations")
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
		Data:      responses,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		PageCount: pageCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(listResp); err != nil {
		log.Printf("[api] failed to encode response: %v", err)
	}
}

// GetViolationStatsHandler returns violation statistics
func GetViolationStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d >= 0 {
			if d > 365 {
				d = 365
			}
			if d == 0 {
				days = 365
			} else {
				days = d
			}
		}
	}

	timeBack := time.Duration(days*24) * time.Hour

	stats, err := internal.GetViolationStats(timeBack)
	if err != nil {
		log.Printf("[api] failed to get stats: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve statistics")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("[api] failed to encode response: %v", err)
	}
}

// respondError sends a standardized error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}
