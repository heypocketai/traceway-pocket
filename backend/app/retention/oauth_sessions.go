package retention

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	traceway "go.tracewayapp.com"
)

const oauthSessionsPruneInterval = 24 * time.Hour

func startOAuthSessionsPrune(ctx context.Context) {
	config.Logf("Starting oauth_sessions prune worker (interval: %s)", oauthSessionsPruneInterval)

	go func() {
		defer traceway.Recover()

		runOAuthSessionsPrune(ctx)

		ticker := time.NewTicker(oauthSessionsPruneInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runOAuthSessionsPrune(ctx)
			}
		}
	}()
}

func runOAuthSessionsPrune(ctx context.Context) {
	if ctx.Err() != nil || db.DB == nil {
		return
	}

	_, err := db.ExecuteTransaction(func(tx *sql.Tx) (int64, error) {
		return repositories.OAuthSessionRepository.PruneExpired(tx, time.Now().UTC())
	})
	if err != nil {
		traceway.CaptureException(fmt.Errorf("retention: prune oauth_sessions failed: %w", err))
	}
}
