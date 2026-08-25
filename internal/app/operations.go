package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/tirecure/internal/model"
)

func (a *App) ValidateMoldDrift(ctx context.Context, moistPct float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	limit := a.cfg.TargetMoistPct + a.cfg.MaxGradientDeltaPct
	if moistPct <= limit {
		return nil
	}
	return fmt.Errorf("moisture: %w", model.ErrMoldDrift)
}

func (a *App) ConfirmBladderHold(ctx context.Context, anchor time.Time) error {
	if a.avgWindow == nil {
		return model.Wrap("app", "window", model.ErrBladderHold)
	}
	if err := a.avgWindow.Require(anchor); err != nil {
		return fmt.Errorf("gradient hold: %w", err)
	}
	return nil
}
