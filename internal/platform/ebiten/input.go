package ebiten

import (
	ebt "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/gentaman/myrogue/internal/app"
)

func PollInput() app.InputState {
	var state app.InputState

	if inpututil.IsKeyJustPressed(ebt.KeyArrowUp) || inpututil.IsKeyJustPressed(ebt.KeyW) {
		state.DirPressed = true
		state.Dir = 0
		state.Up = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyArrowDown) || inpututil.IsKeyJustPressed(ebt.KeyS) {
		state.DirPressed = true
		state.Dir = 1
		state.Down = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebt.KeyA) {
		state.DirPressed = true
		state.Dir = 2
		state.Left = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyArrowRight) || inpututil.IsKeyJustPressed(ebt.KeyD) {
		state.DirPressed = true
		state.Dir = 3
		state.Right = true
	}

	if inpututil.IsKeyJustPressed(ebt.KeyEnter) || inpututil.IsKeyJustPressed(ebt.KeySpace) {
		state.Confirm = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyEscape) {
		state.Cancel = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyX) {
		state.Menu = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyI) {
		state.Inventory = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyL) {
		state.Log = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyM) {
		state.MapView = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyH) {
		state.Help = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyO) {
		state.Options = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyC) {
		state.Credit = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyE) {
		state.CharCreate = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyR) {
		state.Restart = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyZ) {
		state.Skill = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyY) {
		state.Yes = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyN) {
		state.No = true
	}

	if inpututil.IsKeyJustPressed(ebt.KeyF1) {
		state.Debug = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF2) {
		state.DebugGrid = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF3) {
		state.DebugFOV = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF4) {
		state.DebugEntityID = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF5) {
		state.DebugReveal = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF6) {
		state.DebugSlowAnim = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF7) {
		state.DebugPause = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF8) {
		state.DebugStep = true
	}
	if inpututil.IsKeyJustPressed(ebt.KeyF9) {
		state.DebugSkipAnim = true
	}

	return state
}
