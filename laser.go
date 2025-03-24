package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Laser struct {
	sprite Sprite
	usedUp bool
}

func (l *Laser) Update() {
	l.sprite.x += l.sprite.dirX
	l.sprite.y -= l.sprite.dirY
}

func (l Laser) IsOffScreen() bool {
	return l.sprite.x < 0 || l.sprite.x > float64(screenWidth) || l.sprite.y < 0 || l.sprite.y > float64(screenHeight)
}

func (l Laser) GetBounds() image.Rectangle {
	// Create a GeoM transformation matrix
	return l.sprite.GetBounds()
}

func (l *Laser) HandleCollision(c Collidable) {
	_, ok := c.(*Asteroid)
	if ok {
		l.usedUp = true
	}
}

func (l *Laser) createBeam() {
	l.sprite.image = ebiten.NewImage(l.sprite.width, l.sprite.height)
	l.sprite.image.Fill(color.RGBA{255, 0, 0, 255})
}
