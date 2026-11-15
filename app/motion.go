package main

import (
	"math"
	"math/rand"
)

type Behavior uint8

const (
	Walk Behavior = iota
	Run
	Stalk
	Rest
	Pounce
)

func (b Behavior) String() string {
	switch b {
	case Walk:
		return "Walk"
	case Run:
		return "Run"
	case Stalk:
		return "Stalk"
	case Rest:
		return "Rest"
	case Pounce:
		return "Pounce"
	default:
		return "Unknown"
	}
}

type Walker struct {
	X               float64
	Direction       float64 // 1 moves right, -1 moves left
	State           Behavior
	FrameIndex      int
	Phase           float64
	Age             float64
	Duration        float64
	SpeedMultiplier float64
	Paused          bool
	Automatic       bool
}

func NewWalker(x float64) Walker {
	return Walker{X: x, Direction: -1, State: Walk, Duration: 7, SpeedMultiplier: 1, Automatic: true}
}
func (w *Walker) ChooseState(r *rand.Rand) {
	n := r.Intn(100)
	switch {
	case n < 58:
		w.State = Walk
		w.Duration = 6 + float64(r.Intn(4))
	case n < 73:
		w.State = Stalk
		w.Duration = 3.6
	case n < 84:
		w.State = Run
		w.Duration = 2.5
	case n < 96:
		w.State = Rest
		w.Duration = 2.3
	default:
		w.State = Pounce
		w.Duration = 1.1
	}
	w.Age = 0
}
func (w *Walker) SetState(b Behavior) {
	w.State = b
	w.Age = 0
	w.Phase = 0
	w.Automatic = false
	w.Paused = false
	w.Duration = 9999
	if b == Pounce {
		w.Duration = 1.1
	}
}
func (w *Walker) Update(dt, left, right float64, r *rand.Rand) (vertical int) {
	if dt < 0 {
		dt = 0
	}
	if dt > 0.12 {
		dt = 0.12
	}
	if w.Paused {
		return 0
	}
	w.Age += dt
	if w.Age >= w.Duration {
		if w.Automatic {
			w.ChooseState(r)
		} else if w.State == Pounce {
			w.State = Rest
			w.Age = 0
			w.Duration = 9999
		}
	}
	var dx, fps float64
	switch w.State {
	case Walk:
		dx = 105
		fps = 9
	case Run:
		dx = 220
		fps = 15
	case Stalk:
		dx = 42
		fps = 5
	case Rest:
		dx = 0
		fps = 0
	case Pounce:
		dx = 165
		fps = 11
	}
	w.X += dt * dx * w.Direction * w.SpeedMultiplier
	if w.X < left {
		w.X = left
		w.Direction = 1
	}
	if w.X > right {
		w.X = right
		w.Direction = -1
	}
	w.Phase += dt * fps * w.SpeedMultiplier
	if fps > 0 {
		w.FrameIndex = int(math.Floor(w.Phase)) % 8
	} else {
		w.FrameIndex = 0
	}
	if w.State == Pounce {
		vertical = -int(53 * math.Sin(math.Pi*math.Min(w.Age/w.Duration, 1)))
	}
	if w.State == Run {
		vertical = -int(2 * math.Abs(math.Sin(w.Age*18)))
	}
	if w.State == Stalk {
		vertical = 5
	}
	if w.State == Rest {
		vertical = int(1.5 * math.Sin(w.Age*2.6))
	}
	return vertical
}
