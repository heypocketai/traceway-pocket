package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type tableParts struct {
	Table string `json:"table"`
	Parts int64  `json:"parts"`
}

type chError struct {
	Name          string `json:"name"`
	Value         int64  `json:"value"`
	LastErrorTime string `json:"lastErrorTime,omitempty"`
}

type chSnapshot struct {
	Reachable        bool         `json:"reachable"`
	UptimeSec        int64        `json:"uptimeSec"`
	PartsCount       int64        `json:"partsCount"`
	PartsByTable     []tableParts `json:"partsByTable,omitempty"`
	ActiveMerges     int64        `json:"activeMerges"`
	LongestMergeSec  float64      `json:"longestMergeSec"`
	ErrorsRecent     []chError    `json:"errorsRecent,omitempty"`
	MemoryUsageBytes int64        `json:"memoryUsageBytes,omitempty"`
	MemoryTotalBytes int64        `json:"memoryTotalBytes,omitempty"`
}

// healthDeepBody mirrors the backend HealthDeepResponse with its JSON tags. The
// backend uses `chReachable`/`chUptimeSec`; we expose those as `reachable`/
// `uptimeSec` in our embedded snapshot so the bench JSON stays consistent with
// other loadgen fields.
type healthDeepBody struct {
	CHReachable      bool         `json:"chReachable"`
	CHUptimeSec      int64        `json:"chUptimeSec"`
	PartsCount       int64        `json:"partsCount"`
	PartsByTable     []tableParts `json:"partsByTable"`
	ActiveMerges     int64        `json:"activeMerges"`
	LongestMergeSec  float64      `json:"longestMergeSec"`
	ErrorsRecent     []chError    `json:"errorsRecent"`
	MemoryUsageBytes int64        `json:"memoryUsageBytes"`
	MemoryTotalBytes int64        `json:"memoryTotalBytes"`
}

// fetchCHSnapshot pings the backend's /health/deep endpoint and translates the
// payload into a chSnapshot. On any error (timeout, transport, non-2xx, decode)
// it returns a snapshot with Reachable=false and logs to stderrPrefix() — the
// caller does NOT fail the step on a missing snapshot. A 503 from the backend
// (CH unreachable but backend alive) is still parsed and embedded so the JSON
// captures the chReachable=false signal.
func fetchCHSnapshot(ctx context.Context, cfg config, client *http.Client) chSnapshot {
	if cfg.jwt == "" {
		return chSnapshot{}
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, cfg.target+"/api/health/deep", nil)
	if err != nil {
		fmt.Fprintf(stderrPrefix(), "fetchCHSnapshot: build request failed: %v\n", err)
		return chSnapshot{}
	}
	req.Header.Set("Authorization", "Bearer "+cfg.jwt)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(stderrPrefix(), "fetchCHSnapshot: http error: %v\n", err)
		return chSnapshot{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		fmt.Fprintf(stderrPrefix(), "fetchCHSnapshot: unexpected status %d\n", resp.StatusCode)
		return chSnapshot{}
	}

	var body healthDeepBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		fmt.Fprintf(stderrPrefix(), "fetchCHSnapshot: decode failed: %v\n", err)
		return chSnapshot{}
	}

	return chSnapshot{
		Reachable:        body.CHReachable,
		UptimeSec:        body.CHUptimeSec,
		PartsCount:       body.PartsCount,
		PartsByTable:     body.PartsByTable,
		ActiveMerges:     body.ActiveMerges,
		LongestMergeSec:  body.LongestMergeSec,
		ErrorsRecent:     body.ErrorsRecent,
		MemoryUsageBytes: body.MemoryUsageBytes,
		MemoryTotalBytes: body.MemoryTotalBytes,
	}
}
