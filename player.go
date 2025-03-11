package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	character *ebiten.Image
	rad       float64
	x         float64
	y         float64
	dirX      float64
	dirY      float64
	height    int
	width     int
}

func (p *Player) createCharacter() {
	p.character = ebiten.NewImage(p.height, p.width)
}

func (p *Player) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(-float64(p.width)/2, -float64(p.height)/2)
	geo.Rotate(p.rad)

	geo.Translate(p.x, p.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(p.character, op)
}
