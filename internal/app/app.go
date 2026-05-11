package app

import "github.com/gentaman/myrogue/internal/core/content"

type App struct {
	Scene    Scene
	Registry *content.Registry
	Audio    AudioPlayer
}

type AudioPlayer interface {
	PlaySFX(id string)
}

func NewApp(reg *content.Registry, audio AudioPlayer) *App {
	return &App{
		Scene:    NewTitleSceneWithDeps(reg, audio),
		Registry: reg,
		Audio:    audio,
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
