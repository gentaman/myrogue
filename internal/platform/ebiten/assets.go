package ebiten

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"

	ebt "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed data/fonts/mplus1.ttf
var fontData []byte

//go:embed data/images/player.png
var playerPNG []byte

//go:embed data/sfx/Hit-1.wav
var sfxHitData []byte

//go:embed data/sfx/stair_down.wav
var sfxStairDownData []byte

//go:embed data/sfx/stair_up.wav
var sfxStairUpData []byte

//go:embed data/sfx/Coin-3.wav
var sfxCoinData []byte

//go:embed data/enemies.json
var EnemiesJSON []byte

//go:embed data/companions.json
var CompanionsJSON []byte

//go:embed data/player.json
var PlayerJSON []byte

//go:embed data/items.json
var ItemsJSON []byte

//go:embed data/floors.json
var FloorsJSON []byte

//go:embed data/skills.json
var SkillsJSON []byte

var (
	fontFaceSource *text.GoTextFaceSource
	fontFace12     *text.GoTextFace
	fontFace14     *text.GoTextFace
	fontFace20     *text.GoTextFace
	fontFace32     *text.GoTextFace

	audioContext    *audio.Context
	sfxHitPCM       []byte
	sfxStairDownPCM []byte
	sfxStairUpPCM   []byte
	sfxCoinPCM      []byte
	sfxVolume       float64 = 0.5

	spriteImages map[string]*ebt.Image
)

func init() {
	spriteImages = make(map[string]*ebt.Image)
	loadSprite := func(name string, data []byte) {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to decode sprite %s: %v", name, err))
		}
		spriteImages[name] = ebt.NewImageFromImage(img)
	}
	loadSprite("player", playerPNG)

	var err error
	fontFaceSource, err = text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Fatal(err)
	}
	fontFace12 = &text.GoTextFace{Source: fontFaceSource, Size: 12}
	fontFace14 = &text.GoTextFace{Source: fontFaceSource, Size: 14}
	fontFace20 = &text.GoTextFace{Source: fontFaceSource, Size: 20}
	fontFace32 = &text.GoTextFace{Source: fontFaceSource, Size: 32}

	audioContext = audio.NewContext(44100)

	decodeSFX := func(data []byte) []byte {
		s, err := wav.DecodeWithSampleRate(44100, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		pcm, err := io.ReadAll(s)
		if err != nil {
			log.Fatal(err)
		}
		return pcm
	}

	sfxHitPCM = decodeSFX(sfxHitData)
	sfxStairDownPCM = decodeSFX(sfxStairDownData)
	sfxStairUpPCM = decodeSFX(sfxStairUpData)
	sfxCoinPCM = decodeSFX(sfxCoinData)
}
