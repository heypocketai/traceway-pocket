//go:build !pgch

package controllers

import "context"

func fetchCHHealth(_ context.Context) HealthDeepResponse {
	return HealthDeepResponse{CHReachable: false}
}
