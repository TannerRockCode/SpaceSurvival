package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{
	player Player
	lasers []Laser
}

type Laser struct{
	x float64 
	y float64
	rad float64
	dirX float64
	dirY float64
	beam *ebiten.Image
	width int
	height int
}

func (l *Laser) createBeam() {
	l.beam = ebiten.NewImage(l.width, l.height)
	l.beam.Fill(color.RGBA{255, 0, 0, 255})
}

func (l *Laser) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(-float64(l.width)/2, -float64(l.height)/2)
	geo.Rotate(l.rad)
	geo.Translate(l.x, l.y)
	op := &ebiten.DrawImageOptions{GeoM:geo}
	screen.DrawImage(l.beam, op)
}

func (g *Game) DrawLasers(screen *ebiten.Image) {
	for _, l := range g.lasers {
		l.Draw(screen)
	}
}

func (l Laser) IsOffScreen() bool {
	return l.x < 0 || l.x > 960 || l.y < 0 || l.y > 540 
}

func (g *Game) CleanLasers() {
	i := 0
	for _, v := range g.lasers {
		if !v.IsOffScreen() {
			g.lasers[i] = v
			i++
		}
	}
	g.lasers = g.lasers[:i]
}

func (g *Game) MovePlayer() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.rad -= 0.1
		g.player.dirX = math.Sin(g.player.rad)
		g.player.dirY = math.Cos(g.player.rad)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.rad += 0.1
		g.player.dirX = math.Sin(g.player.rad)
		g.player.dirY = math.Cos(g.player.rad)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.y -= g.player.dirY
		g.player.x += g.player.dirX
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.y += g.player.dirY
		g.player.x -= g.player.dirX
	}
}

func (g *Game) PlayerShoot() {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		laser := Laser {
			x : g.player.x,
			y : g.player.y,
			rad : g.player.rad,
			dirX : g.player.dirX * 1.3,
			dirY : g.player.dirY * 1.3,
			width : 1,
			height : 3,
		}
		laser.createBeam()
		g.lasers = append(g.lasers, laser)
	}
}

func (g *Game) Update() error {
	g.LaserShoot()
	g.MoveLasers()
	g.CleanLasers()
	g.MovePlayer()
	g.PlayerShoot()

	return nil
}

func (g *Game) MoveLasers() {
	for i, _ := range g.lasers {
		g.lasers[i].Update()
	}
}

func (g *Game) LaserShoot() {
	currLaser := Laser {
		x: 0,
		y: 0,
	}
	if len(g.lasers) > 0 {
		currLaser = g.lasers[0]
	}
	currLaser.x += 1
	currLaser.y += 1
}

func (l *Laser) Update() {
	l.x += l.dirX
	l.y -= l.dirY
	
}

func (g *Game) Draw(screen *ebiten.Image) {
	firstLaser := Laser {
		x : 0,
		y : 0,
		rad : 0,
		dirX : 0,
		dirY : 0,
		width : 0,
		height : 0,
	}
	if len(g.lasers) > 0 {
		firstLaser = g.lasers[0];	
	} 
	
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Ticks Per Second: %0.2f. x: %f y: %f", ebiten.ActualTPS(), firstLaser.x, firstLaser.y))
	g.player.Draw(screen)
	g.DrawLasers(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}

func main() {
	fmt.Printf("I am drawing!")
	ebiten.SetWindowSize(1920, 1080)
	player := Player{
		height : 10,
		width : 10,
		dirX: 0,
		dirY: 1,
		x : 475,
		y : 265,
	}
	lasers := make([]Laser, 0, 10)
	player.createCharacter()
	player.character.Fill(color.White)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(&Game{player, lasers}); err != nil {
		log.Fatal(err)
	}

}
