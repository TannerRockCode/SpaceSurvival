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
	p.sprite.Draw(screen)
}
