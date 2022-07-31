package utils

import "time"

type (
	Debouncer struct {
		Delay time.Duration
		Func  func()

		timer *time.Timer
	}
)

func (d *Debouncer) Trigger() {
	if d.timer == nil {
		d.timer = time.AfterFunc(d.Delay, d.fire)
	} else {
		d.timer.Reset(d.Delay)
	}
}

func (d Debouncer) Stop() {
	if d.timer == nil {
		return
	}
	if d.timer.Stop() {
		<-d.timer.C
	}
}

func (d Debouncer) fire() {
	if d.Func != nil {
		d.Func()
	}
}
