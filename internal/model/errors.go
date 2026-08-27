package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("tirecure: invalid identifier")
	ErrNotFound        = errors.New("tirecure: entity not found")
	ErrConflict        = errors.New("tirecure: state conflict")
	ErrInterlock       = errors.New("tirecure: interlock denied")
	ErrMoistureHold    = errors.New("tirecure: moisture hold active")
	ErrAirflowSetpoint = errors.New("tirecure: airflow setpoint violation")
	ErrFanFault        = errors.New("tirecure: fan fault")
	ErrScheduleEmpty   = errors.New("tirecure: schedule empty")
	ErrGradient        = errors.New("tirecure: moisture gradient violation")
	ErrMoldDrift   = errors.New("tirecure: moisture drift exceeded")
	ErrCureTrip    = errors.New("tirecure: heat overtemperature")
	ErrBladderHold    = errors.New("tirecure: gradient hold not satisfied")
	ErrContextCanceled = errors.New("tirecure: operation canceled")
)

type DomainError struct {
	Op   string
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("tirecure %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("tirecure %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool   { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }
