package main

import (
	"image"
	"math"
)

type Crystal struct {
	sprite   Sprite
	absorbed bool
}

func (c Crystal) GetBounds() image.Rectangle {
	return c.sprite.GetBounds()
}

func (c *Crystal) Update(p Player) {
	playerXDist := (p.sprite.x - float64(c.sprite.x))
	playerYDist := (p.sprite.y - float64(c.sprite.y))
	crystalDividend := math.Abs(playerXDist) + math.Abs(playerYDist)/crystalAcceleration
	forceX := playerXDist / crystalDividend
	forceY := playerYDist / crystalDividend
	c.sprite.dirX = c.sprite.dirX*.9 + forceX
	c.sprite.dirY = c.sprite.dirY*.9 + forceY
	c.sprite.x += c.sprite.dirX
	c.sprite.y += c.sprite.dirY

	if c.GetBounds().Overlaps(p.GetBounds()) {
		c.absorbed = true
	}
}
