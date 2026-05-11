package debug

import "github.com/gentaman/myrogue/internal/clock"

type State struct {
	Enabled      bool
	ShowHUD      bool
	ShowFOV      bool
	ShowGrid     bool
	ShowEntityID bool
	RevealMap    bool
	Clock        *clock.ScaledClock
}

func NewState() *State {
	return &State{
		Clock: clock.New(),
	}
}

func (s *State) Toggle() {
	s.ShowHUD = !s.ShowHUD
	s.Enabled = s.ShowHUD
}
