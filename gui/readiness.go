package gui

import "time"

type GuiReadinessCheck struct {
	t   time.Time
	Ack chan interface{}
}

func (e GuiReadinessCheck) When() time.Time {
	return e.t
}

func NewGuiReadinessCheck(ack chan interface{}) *GuiReadinessCheck {
	return &GuiReadinessCheck{
		t:   time.Now(),
		Ack: ack,
	}
}
