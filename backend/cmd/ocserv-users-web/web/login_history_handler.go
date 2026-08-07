package web

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"ocserv-users/internal"
)

// loginHistoryWebHandler serves the main login history web page
func loginHistoryWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "login-history.html", nil)
}

// GetLoginHistoryHandler handles API requests for login history data
func GetLoginHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := internal.LoginHistoryQuery{
		Limit: internal.LoginHistoryDefaultLimit,
	}

	// Time range parameters
	timeBackStr := r.URL.Query().Get("time_back")
	if timeBackStr != "" {
		timeBack, err := strconv.Atoi(timeBackStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid time_back parameter")
			return
		}
		query.TimeBack = time.Duration(timeBack) * 24 * time.Hour
	}

	startTimeStr := r.URL.Query().Get("start_time")
	if startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid start_time parameter, expected RFC3339 format")
			return
		}
		query.StartTime = startTime
		query.UseCustomRange = true
	}

	endTimeStr := r.URL.Query().Get("end_time")
	if endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid end_time parameter, expected RFC3339 format")
			return
		}
		query.EndTime = endTime
		query.UseCustomRange = true
	}

	// Username filter
	username := r.URL.Query().Get("username")
	query.Username = username

	// Pagination
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		query.Limit = limit
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		query.Offset = offset
	}

	// Active only filter
	activeOnlyStr := r.URL.Query().Get("active_only")
	if activeOnlyStr == "true" {
		query.ShowActiveOnly = true
	}

	// Get login history records
	records, totalCount, err := internal.GetLoginHistory(query)
	if err != nil {
		log.Printf("[login_history] failed to get login history: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to retrieve login history")
		return
	}

	// Format response
	response := map[string]interface{}{
		"records":     records,
		"total_count": totalCount,
		"limit":       query.Limit,
		"offset":      query.Offset,
	}

	sendSuccess(w, "login history retrieved successfully", response)
}
