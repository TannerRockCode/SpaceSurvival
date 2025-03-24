package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	x      float64
	y      float64
	rad    float64
	dirX   float64
	dirY   float64
	image  *ebiten.Image
	width  int
	height int
}

func (s *Sprite) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	if s.rad != 0 {
		geo.Translate(-float64(s.width)/2, -float64(s.height)/2)
		geo.Rotate(s.rad)
	}
	geo.Translate(s.x, s.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(s.image, op)
}

func (s Sprite) GetBounds() image.Rectangle {
	if s.rad == 0 {
		return image.Rect(int(s.x), int(s.y), int(s.x+float64(s.width)), int(s.y+float64(s.height)))
	}

	geo := ebiten.GeoM{}
	geo.Translate(-float64(s.width)/2, -float64(s.height)/2)
	geo.Rotate(s.rad)
	geo.Translate(s.x, s.y)

	// Get the transformed corners of the laser
	x0, y0 := geo.Apply(0, 0)
	x1, y1 := geo.Apply(float64(s.width), 0)
	x2, y2 := geo.Apply(0, float64(s.height))
	x3, y3 := geo.Apply(float64(s.width), float64(s.height))

	// Find the bounding box that contains all the transformed points
	minX := math.Min(math.Min(x0, x1), math.Min(x2, x3))
	maxX := math.Max(math.Max(x0, x1), math.Max(x2, x3))
	minY := math.Min(math.Min(y0, y1), math.Min(y2, y3))
	maxY := math.Max(math.Max(y0, y1), math.Max(y2, y3))

	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}
