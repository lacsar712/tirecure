package fsm

import (
	"context"
	"fmt"

	"github.com/lacsar712/tirecure/internal/model"
)

var ErrIllegalDryTransition = fmt.Errorf("illegal dry transition")

type CureFSM struct {
	id    model.TowerID
	state model.DryState
	hooks *DryHookChain
}

func NewCureFSM(id model.TowerID, effect func(context.Context, model.TowerID, model.DryState, model.DryState) error) *CureFSM {
	_ = effect
	return &CureFSM{id: id, state: model.DryIdle, hooks: NewDryHookChain()}
}

func (f *CureFSM) Hooks() *DryHookChain { return f.hooks }

func (f *CureFSM) State() model.DryState { return f.state }

func (f *CureFSM) Dispatch(ctx context.Context, event string) (model.DryState, error) {
	next, ok := allowedDry(f.state, event)
	if !ok {
		// Illegal transition: state is unchanged, so the after hooks (which drive
		// the execution side / valve pulses) must NOT run. Mirrors TowerFSM and
		// FanFSM, which bail before invoking their side effect on rejection.
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalDryTransition)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func allowedDry(from model.DryState, event string) (model.DryState, bool) {
	switch from {
	case model.DryIdle:
		if event == "arm_heat" {
			return model.DryHeating, true
		}
	case model.DryHeating:
		if event == "hold" {
			return model.DryHold, true
		}
	case model.DryHold:
		if event == "cool" {
			return model.DryCool, true
		}
	case model.DryCool:
		if event == "done" {
			return model.DryIdle, true
		}
	}
	return from, false
}
