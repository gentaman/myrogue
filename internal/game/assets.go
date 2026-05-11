package game

import (
	"bytes"
	_ "embed"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/fonts/mplus1.ttf
var fontData []byte

//go:embed assets/sfx/Hit-1.wav
var sfxHitData []byte

//go:embed assets/sfx/stair_down.wav
var sfxStairDownData []byte

//go:embed assets/sfx/stair_up.wav
var sfxStairUpData []byte

//go:embed assets/sfx/Coin-3.wav
var sfxCoinData []byte

//go:embed assets/credits.txt
var creditsText string

//go:embed assets/enemies.json
var enemiesJSON []byte

//go:embed assets/companions.json
var companionsJSON []byte

//go:embed assets/player.json
var playerJSON []byte

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
)

func init() {
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

	initActorDefs()
}
