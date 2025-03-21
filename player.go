package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	sprite Sprite
}

func (p *Player) createCharacter() {
	p.sprite.image = ebiten.NewImage(p.sprite.height, p.sprite.width)
}

func (p *Player) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(-float64(p.sprite.width)/2, -float64(p.sprite.height)/2)
	geo.Rotate(p.sprite.rad)

	geo.Translate(p.sprite.x, p.sprite.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(p.sprite.image, op)
}
