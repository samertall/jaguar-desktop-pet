package main

import (
	"math/rand"
	"testing"
)

func TestWalkerMovesAndFramesChange(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	w := NewWalker(200)
	w.Duration = 100
	x := w.X
	first := w.FrameIndex
	for i := 0; i < 20; i++ {
		w.Update(0.04, 0, 1000, r)
	}
	if w.X >= x {
		t.Fatalf("walked wrong direction x=%v x0=%v", w.X, x)
	}
	if w.FrameIndex == first {
		t.Fatalf("animation frame did not advance")
	}
}
func TestPauseFreezesAndEdgeTurns(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	w := NewWalker(100)
	w.Paused = true
	for i := 0; i < 20; i++ {
		w.Update(.04, 0, 600, r)
	}
	if w.X != 100 || w.FrameIndex != 0 {
		t.Fatal("pause did not freeze")
	}
	w.Paused = false
	w.X = 0
	w.Update(.04, 0, 600, r)
	if w.Direction != 1 || w.X != 0 {
		t.Fatal("left edge did not turn")
	}
}
func TestFasterRun(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	a := NewWalker(500)
	b := NewWalker(500)
	a.Duration = 100
	b.Duration = 100
	b.SetState(Run)
	a.Update(.08, 0, 2000, r)
	b.Update(.08, 0, 2000, r)
	if !(b.X < a.X) {
		t.Fatal("run did not move faster")
	}
}
