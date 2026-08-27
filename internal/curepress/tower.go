package curepress

import (
	"context"
	"time"

	"github.com/lacsar712/tirecure/internal/airflow"
	"github.com/lacsar712/tirecure/internal/clock"
	"github.com/lacsar712/tirecure/internal/config"
	"github.com/lacsar712/tirecure/internal/model"
	"github.com/lacsar712/tirecure/internal/moldtemp"
	"github.com/lacsar712/tirecure/internal/store"
)

type ZoneTable struct {
	tower  model.TowerID
	zones  []model.ZoneAssignment
	plenum model.PlenumID
}

func NewZoneTable(tower model.TowerID, count int, plenum model.PlenumID, cfg config.Config) (*ZoneTable, error) {
	if count <= 0 {
		return nil, model.Wrap("zone_table", "count", model.ErrInvalidID)
	}
	z := &ZoneTable{tower: tower, plenum: plenum}
	for i := 0; i < count; i++ {
		zoneID, err := model.ParseZoneID(tower, i)
		if err != nil {
			return nil, err
		}
		z.zones = append(z.zones, model.ZoneAssignment{
			Zone: zoneID, Plenum: plenum, Enabled: true,
			Setpoint: model.AirflowSetpoint{CubicMetersPerHour: cfg.DefaultAirflowCMH, TolerancePct: cfg.AirflowTolerancePct},
			TargetMoist: cfg.TargetMoistPct,
		})
	}
	return z, nil
}

func (z *ZoneTable) Zones() []model.ZoneAssignment {
	out := make([]model.ZoneAssignment, len(z.zones))
	copy(out, z.zones)
	return out
}

func (z *ZoneTable) EnabledCount() int {
	n := 0
	for _, zone := range z.zones {
		if zone.Enabled {
			n++
		}
	}
	return n
}

func (z *ZoneTable) UpdateMoisture(zone model.ZoneID, pct float64) {
	for i := range z.zones {
		if z.zones[i].Zone == zone {
			z.zones[i].ActualMoist = pct
			return
		}
	}
}

func (z *ZoneTable) UpdateFlow(zone model.ZoneID, cmh float64) {
	for i := range z.zones {
		if z.zones[i].Zone == zone {
			z.zones[i].LastFlow = cmh
			return
		}
	}
}

type PressPlant struct {
	cfg       config.Config
	clk       clock.Clock
	mem       *store.Memory
	plenums   *airflow.PlenumTable
	fans      *airflow.FanBank
	sensors   *moldtemp.SensorBank
	gradient  *moldtemp.GradientController
	profile   *moldtemp.ProfileManager
	holdStart time.Time
	holdDur   time.Duration
}

func NewPressPlant(cfg config.Config, clk clock.Clock, mem *store.Memory) *PressPlant {
	return &PressPlant{
		cfg: cfg, clk: clk, mem: mem,
		plenums:  airflow.NewPlenumTable(),
		fans:     airflow.NewFanBank(),
		sensors:  moldtemp.NewSensorBank(),
		gradient: moldtemp.NewGradientController(cfg.MaxGradientDeltaPct),
	}
}

func (p *PressPlant) Plenums() *airflow.PlenumTable { return p.plenums }

func (p *PressPlant) Fans() *airflow.FanBank { return p.fans }

func (p *PressPlant) BindAirflow(plenum model.PlenumID, sp model.AirflowSetpoint) {
	if pl, ok := p.plenums.Get(plenum); ok {
		pl.BindSetpoint(sp)
	}
}

func (p *PressPlant) Coordinator() *Coordinator { return NewCoordinator(p.cfg, p.clk, p.fans) }

func (p *PressPlant) PrimePlenum(ctx context.Context, plenum model.PlenumID) error {
	pl, ok := p.plenums.Get(plenum)
	if !ok {
		return model.Wrap("tower_plant", "plenum", model.ErrNotFound)
	}
	dur := time.Duration(p.cfg.PlenumPrimeSec) * time.Second
	return pl.Prime(ctx, p.clk, dur)
}

func (p *PressPlant) ObserveFlow(plenum model.PlenumID, cmh float64) {
	if pl, ok := p.plenums.Get(plenum); ok {
		pl.ObserveFlow(cmh)
	}
}

func (p *PressPlant) ValidateFlows(ctx context.Context) error {
	return p.plenums.ValidateAll()
}

func (p *PressPlant) RegisterSensor(sensor *moldtemp.Sensor) {
	p.sensors.Register(sensor)
}

func (p *PressPlant) ObserveMoisture(zone model.ZoneID, pct float64) error {
	_, err := p.sensors.ObserveZone(zone, pct, p.clk.Now())
	return err
}

func (p *PressPlant) ValidateGradient() error {
	return p.gradient.Validate(p.sensors.Readings())
}

func (p *PressPlant) InitProfile(zones []model.ZoneID, targets []float64) error {
	pm, err := moldtemp.NewProfileManager(zones, targets)
	if err != nil {
		return err
	}
	p.profile = pm
	return nil
}

func (p *PressPlant) ArmMoistureHold(start time.Time, duration time.Duration, targetPct float64) {
	if p.profile == nil {
		return
	}
	p.profile.Window(start, duration, targetPct)
	p.holdStart = start
	p.holdDur = duration
}

func (p *PressPlant) HoldActive() bool {
	if p.profile == nil {
		return false
	}
	return p.profile.HoldActive(p.clk.Now)
}

func (p *PressPlant) ReleaseHold() {
	if p.profile != nil {
		p.profile.ReleaseHold()
	}
}

func (p *PressPlant) AtTarget(tolerance float64) bool {
	if p.profile == nil {
		return false
	}
	return p.profile.AllAtTarget(p.sensors.Readings(), tolerance)
}

func (p *PressPlant) GradientDelta() float64 {
	return p.gradient.Delta(p.sensors.Readings())
}

func (p *PressPlant) GradientDeltaFor(readings []model.MoistureReading) float64 {
	return p.gradient.Delta(readings)
}

func (p *PressPlant) SensorReadings() []model.MoistureReading {
	return p.sensors.Readings()
}

func (p *PressPlant) Profile() *moldtemp.ProfileManager {
	return p.profile
}

func (p *PressPlant) HoldDuration() time.Duration {
	if p.holdDur > 0 {
		return p.holdDur
	}
	return time.Minute
}
