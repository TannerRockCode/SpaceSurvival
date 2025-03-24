package main

import (
	"image"
)

type Asteroid struct {
	sprite    Sprite
	destroyed bool
}

func (a Asteroid) IsOffScreen() bool {
	return a.sprite.x < 0 || a.sprite.x > float64(screenWidth) || a.sprite.y < 0 || a.sprite.y > float64(screenHeight)
}

func (a Asteroid) GetBounds() image.Rectangle {
	return a.sprite.GetBounds()
}

func (a *Asteroid) Update() {
	a.sprite.x += a.sprite.dirX
	a.sprite.y += a.sprite.dirY
}

func (a *Asteroid) HandleCollision(c Collidable) {
	l, ok := c.(*Laser)
	if ok {
		a.sprite.dirX = l.sprite.dirX
		a.sprite.dirY = l.sprite.dirY
		a.destroyed = true
	}
}
