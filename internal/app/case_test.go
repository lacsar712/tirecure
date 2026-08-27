package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/tirecure/internal/config"
	"github.com/lacsar712/tirecure/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	err = a.HandleCureTrip(context.Background(), model.TowerID(a.cfg.TowerID), 95.0)
	if err == nil {
		t.Fatal("expected heat overtemperature error")
	}
	if !errors.Is(err, model.ErrCureTrip) {
		t.Fatalf("expected ErrCureTrip, got %v", err)
	}
}
