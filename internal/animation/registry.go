package animation

type VisualType int

const (
	VisualProjectile VisualType = iota
	VisualFlash
)

type VisualDef struct {
	Type         VisualType
	TotalFrames  int
	DefaultColor string
}

var Registry = map[string]VisualDef{
	"fireball": {
		Type:         VisualProjectile,
		TotalFrames:  15,
		DefaultColor: "#FF4500",
	},
	"magic_missile": {
		Type:         VisualProjectile,
		TotalFrames:  12,
		DefaultColor: "#64B4FF",
	},
	"holy_light": {
		Type:         VisualFlash,
		TotalFrames:  10,
		DefaultColor: "#FFFFE0",
	},
	"thunder": {
		Type:         VisualProjectile,
		TotalFrames:  10,
		DefaultColor: "#00FFFF",
	},
	"heal_flash": {
		Type:         VisualFlash,
		TotalFrames:  20,
		DefaultColor: "#64FF96",
	},
}
