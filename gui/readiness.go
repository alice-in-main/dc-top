package gui

import "time"

type ReadinessCheck struct {
	t   time.Time
	Ack chan interface{}
}

func (e ReadinessCheck) When() time.Time {
	return e.t
}

func NewReadinessCheck(ack chan interface{}) *ReadinessCheck {
	return &ReadinessCheck{
		t:   time.Now(),
		Ack: ack,
	}
}
