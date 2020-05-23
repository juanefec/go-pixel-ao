package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type ConfigHud struct {
	open   bool
	canvas pixelgl.Canvas
}

type Button struct {
	Rect pixel.Rect
	Pos  pixel.Vec
}

func NewConfigHud() { //*ConfigHud {
	//canvas := pixelgl.NewCanvas(pixel.R(0, 0, 0, 0))
}
