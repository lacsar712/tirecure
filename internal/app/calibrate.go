package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/tirecure/internal/model"
)

// CalibrateProbe allows acceptance tests to inject feed calibration faults.
var CalibrateProbe func(ctx context.Context) error

const feedTempLimitC = 55.0

func (a *App) CalibrateFeed(ctx context.Context, tower model.TowerID, holder string) error {
	if err := a.feedLeases.Require(tower, holder, 30*time.Second); err != nil {
		return err
	}
	// The feed lease marks the bladder channel occupied for the duration of
	// calibration. Release it on every exit so a probe fault cannot leave the
	// occupancy bit latched: a held lease would otherwise block the downstream
	// pressurize step (the TTL is moot here, as the process clock does not
	// advance between the failed calibration and the next-day pressurize).
	defer a.feedLeases.ReleaseHolder(tower, holder)
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	return nil
}
