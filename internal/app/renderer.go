package app

type Color [4]uint8

type Renderer interface {
	ScreenSize() (int, int)
	Clear(r, g, b, a uint8)
	DrawRect(x, y, w, h int, clr Color)
	DrawSprite(name string, frame, dir int, x, y float64)
	DrawText(text string, size int, x, y int, clr Color, center bool)
	DrawPanel(x, y, w, h int, bg, border Color)
	MeasureText(text string, size int) int
}

type InputState struct {
	DirPressed    bool
	Dir           int // 0=up,1=down,2=left,3=right
	Confirm       bool
	Cancel        bool
	Menu          bool
	Inventory     bool
	Log           bool
	MapView       bool
	Help          bool
	Options       bool
	Credit        bool
	CharCreate    bool
	Restart       bool
	Yes           bool
	No            bool
	Up            bool
	Down          bool
	Left          bool
	Right         bool
	Debug         bool
	DebugFOV      bool
	DebugGrid     bool
	DebugEntityID bool
	DebugReveal   bool
	DebugSlowAnim bool
	DebugFastAnim bool
	DebugPause    bool
	DebugStep     bool
	DebugSkipAnim bool
}
