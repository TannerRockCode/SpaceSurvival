package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth = 960
	screenHeight = 540
	asteroidSpeed = 0.9
) 


//var startTime = time.Now()
var lastEnemySpawn = time.Now()

type Game struct{
	player Player
	lasers []Laser
	asteroids []Asteroid
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

type Asteroid struct{
	x float64
	y float64
	dirX float64
	dirY float64
	sprite *ebiten.Image
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

func (a *Asteroid) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(a.x, a.y)
	op := &ebiten.DrawImageOptions{GeoM:geo}
	screen.DrawImage(a.sprite, op)
}

func (g *Game) DrawAsteroids(screen *ebiten.Image) {
	for _, a := range g.asteroids {
		a.Draw(screen)
	}
}

func (l Laser) IsOffScreen() bool {
	return l.x < 0 || l.x > float64(screenWidth) || l.y < 0 || l.y > float64(screenHeight) 
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
	g.SpawnAsteroid()
	g.MoveAsteroids()

	return nil
}

func (g *Game) SpawnAsteroid() {
	if time.Since(lastEnemySpawn) > 1 * time.Second {
		lastEnemySpawn = time.Now()
		//random border location
		xLocation := 0
		yLocation := 0

		isOnYAxis := rand.Intn(2) == 1
		if isOnYAxis {
			isXMin := rand.Intn(2) == 1
			if isXMin {
				xLocation = 0
			} else {
				xLocation = screenWidth
			}
			yLocation = rand.Intn(screenHeight + 1)
		} else {
			isYMin := rand.Intn(2) == 1
			if isYMin {
				yLocation = 0
			} else {
				yLocation = screenHeight
			}
			xLocation = rand.Intn(screenWidth + 1)
		}
		
		playerXDist := (g.player.x - float64(xLocation))
		playerYDist := (g.player.y - float64(yLocation)) 
		asteroidDividend := (playerXDist + playerYDist) / asteroidSpeed
		asteroidWidth := rand.Intn(40) + 10
		asteroidHeight := rand.Intn(40) + 10

		asteroid := Asteroid{
			x: float64(xLocation), 
			y: float64(yLocation), 
			dirX: (playerXDist / asteroidDividend), 
			dirY: (playerYDist / asteroidDividend), 
			sprite: ebiten.NewImage(asteroidWidth, asteroidHeight),
			width: asteroidWidth,
			height: asteroidHeight,
		}
		asteroid.sprite.Fill(color.RGBA{177, 10, 75, 255})
		g.asteroids = append(g.asteroids, asteroid)
	}
}

func (g *Game) MoveAsteroids() {
	for i, _ := range g.asteroids {
		g.asteroids[i].Update()
	}
}

func (a *Asteroid) Update() {
	a.x += a.dirX
	a.y += a.dirY
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Ticks Per Second: %0.2f", ebiten.ActualTPS()))
	g.player.Draw(screen)
	g.DrawLasers(screen)
	g.DrawAsteroids(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return screenWidth, screenHeight
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
	lasers := make([]Laser, 0, 500)
	asteroids := make([]Asteroid, 0, 25)
	player.createCharacter()
	player.character.Fill(color.White)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(&Game{player, lasers, asteroids}); err != nil {
		log.Fatal(err)
	}

}
