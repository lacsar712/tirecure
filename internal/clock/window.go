package clock

import (
	"time"

	"github.com/lacsar712/tirecure/internal/model"
)

type CureWindow struct {
	clk      Clock
	duration time.Duration
}

func NewCureWindow(clk Clock, duration time.Duration) *CureWindow {
	if duration <= 0 {
		duration = 2 * time.Minute
	}
	return &CureWindow{clk: clk, duration: duration}
}

func (w *CureWindow) Active(anchor time.Time) bool {
	return WindowElapsed(w.clk, anchor, w.duration)
}

func (w *CureWindow) Require(anchor time.Time) error {
	if w.Active(anchor) {
		return nil
	}
	return model.ErrBladderHold
}
