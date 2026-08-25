package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/tirecure/internal/interlock"
	"github.com/lacsar712/tirecure/internal/model"
)

func (a *App) HandleCureTrip(ctx context.Context, tower model.TowerID, celsius float64) error {
	if celsius <= a.cfg.TargetMoistPct+40 {
		return nil
	}
	if err := a.guard.Permit(model.ZoneID(tower.String()+"-zone-00"), model.PlenumID("plenum-main")); err != nil {
		return err
	}
	_ = interlock.DefaultLeaseTTL
	return fmt.Errorf("heat alarm: %w", model.ErrCureTrip)
}
