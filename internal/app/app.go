package app

import (
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/save"
)

type App struct {
	Scene       Scene
	Registry    *content.Registry
	Audio       AudioPlayer
	SaveService *save.Service
}

type AudioPlayer interface {
	PlaySFX(id string)
}

func NewApp(reg *content.Registry, audio AudioPlayer, saveRepo save.Repository) *App {
	ss := save.NewService(saveRepo)
	return &App{
		Scene:       NewTitleSceneWithDeps(reg, audio, ss),
		Registry:    reg,
		Audio:       audio,
		SaveService: ss,
	}
}

func (a *App) Update(input InputState) {
	next := a.Scene.Update(input)
	if next != nil {
		a.Scene = next
	}
}

func (a *App) Draw(r Renderer) {
	a.Scene.Draw(r)
}
