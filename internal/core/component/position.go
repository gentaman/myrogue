package component

type Dir int

const (
	DirUp Dir = iota
	DirDown
	DirLeft
	DirRight
)

func (d Dir) Delta() (int, int) {
	switch d {
	case DirUp:
		return 0, -1
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	default:
		return 1, 0
	}
}

func DirFromDelta(dx, dy int) Dir {
	switch {
	case dx > 0:
		return DirRight
	case dx < 0:
		return DirLeft
	case dy > 0:
		return DirDown
	default:
		return DirUp
	}
}

type Position struct {
	X, Y int
	Dir  Dir
}
