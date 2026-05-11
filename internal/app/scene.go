package app

type Scene interface {
	Update(input InputState) Scene
	Draw(r Renderer)
}
