package main

import (
	"math/rand"
	"testing"
)

func TestRightEdgeDirectionAndBounds(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	w := NewWalker(980)
	w.Direction = 1
	w.Duration = 100
	w.Update(.12, 0, 990, r)
	if w.X > 990 || w.Direction != -1 {
		t.Fatalf("right edge turn failed x=%v dir=%v", w.X, w.Direction)
	}
}
func TestRestDoesNotMoveHorizontally(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	w := NewWalker(200)
	w.SetState(Rest)
	for i := 0; i < 20; i++ {
		w.Update(.04, 0, 1000, r)
	}
	if w.X != 200 || w.FrameIndex != 0 {
		t.Fatalf("rest movement changed x=%v frame=%v", w.X, w.FrameIndex)
	}
}
func TestPounceReturnsToRestInManualMode(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	w := NewWalker(200)
	w.SetState(Pounce)
	for i := 0; i < 40; i++ {
		w.Update(.04, 0, 1000, r)
	}
	if w.State != Rest {
		t.Fatalf("pounce did not return to rest: %s", w.State)
	}
}
