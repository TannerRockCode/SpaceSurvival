package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth       = 960
	screenHeight      = 540
	asteroidSpeed     = 0.9
	collisionGridSize = 5
)

var logger *log.Logger

var lastEnemySpawn = time.Now()

type Game struct {
	player       Player
	lasers       []Laser
	asteroids    []Asteroid
	collisionMap CollisionMap
}

type CollisionMap struct {
	grid map[string][]Collidable
}

type Laser struct {
	x      float64
	y      float64
	rad    float64
	dirX   float64
	dirY   float64
	beam   *ebiten.Image
	width  int
	height int
	usedUp bool
}

type Asteroid struct {
	x      float64
	y      float64
	dirX   float64
	dirY   float64
	sprite *ebiten.Image
	width  int
	height int
}

func (c *CollisionMap) Clear() {
	for i := range c.grid {
		clear(c.grid[i])
	}
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
	op := &ebiten.DrawImageOptions{GeoM: geo}
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
	op := &ebiten.DrawImageOptions{GeoM: geo}
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
		if !v.IsOffScreen() && !v.usedUp {
			g.lasers[i] = v
			i++
		}
	}
	g.lasers = g.lasers[:i]
}

func (a Asteroid) IsOffScreen() bool {
	return a.x < 0 || a.x > float64(screenWidth) || a.y < 0 || a.y > float64(screenHeight)
}

func (g *Game) CleanAsteroids() {
	i := 0
	for _, a := range g.asteroids {
		if !a.IsOffScreen() {
			g.asteroids[i] = a
			i++
		}
	}
	g.asteroids = g.asteroids[:i]
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
		laser := Laser{
			x:      g.player.x,
			y:      g.player.y,
			rad:    g.player.rad,
			dirX:   g.player.dirX * 1.3,
			dirY:   g.player.dirY * 1.3,
			width:  1,
			height: 3,
			usedUp: false,
		}
		laser.createBeam()
		g.lasers = append(g.lasers, laser)
	}
}

type Collidable interface {
	//registerWithCollisionMap()
	//detectCollision(Collidable, Collidable) bool
	GetBounds() image.Rectangle
	HandleCollision(Collidable)
}

func (l Laser) GetBounds() image.Rectangle {
	// Create a GeoM transformation matrix
	geo := ebiten.GeoM{}
	geo.Translate(-float64(l.width)/2, -float64(l.height)/2)
	geo.Rotate(l.rad)
	geo.Translate(l.x, l.y)

	// Get the transformed corners of the laser
	x0, y0 := geo.Apply(0, 0)
	x1, y1 := geo.Apply(float64(l.width), 0)
	x2, y2 := geo.Apply(0, float64(l.height))
	x3, y3 := geo.Apply(float64(l.width), float64(l.height))

	// Find the bounding box that contains all the transformed points
	minX := math.Min(math.Min(x0, x1), math.Min(x2, x3))
	maxX := math.Max(math.Max(x0, x1), math.Max(x2, x3))
	minY := math.Min(math.Min(y0, y1), math.Min(y2, y3))
	maxY := math.Max(math.Max(y0, y1), math.Max(y2, y3))

	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

func (a Asteroid) GetBounds() image.Rectangle {
	return image.Rect(int(a.x), int(a.y), int(a.x+float64(a.width)), int(a.y+float64(a.height)))
}

func detectCollision(c1, c2 Collidable) bool {
	b1 := c1.GetBounds()
	b2 := c2.GetBounds()
	return b1.Overlaps(b2)
}

func (l *Laser) HandleCollision(c Collidable) {
	_, ok := c.(*Asteroid)
	if ok {
		l.usedUp = true
	}
}

func (a *Asteroid) HandleCollision(c Collidable) {
	_, ok := c.(*Laser)
	if ok {
		a.sprite.Fill(color.RGBA{50, 0, 0, 255})
	}
}

//for each value in key, for each other value in key if(detectCollision(key1, key2) then key1.handleCollision(key2) key2.handleCollision(key1)

// hashmap grid 96 columns 54 rows - key, which is a combo of the column and row - value is a slice of collidable objects
// register. For every object in a key, detect collision.
//for every value in a key, check to see if that value is colliding with the other values
//c.detectCollision(c2)

func (g *Game) Update() error {
	g.clearCollisionMap()
	g.LaserShoot()
	g.MoveLasers()
	g.MoveAsteroids()
	g.registerCollidables()
	g.handleCollisions()
	g.MovePlayer()
	g.PlayerShoot()
	g.SpawnAsteroid()
	g.CleanLasers()
	g.CleanAsteroids()

	return nil
}

func (g *Game) clearCollisionMap() {
	for i := range g.collisionMap.grid {
		g.collisionMap.grid[i] = make([]Collidable, 0, 100)
	}
}

func (g *Game) registerCollidables() {
	for i := range g.asteroids {
		g.registerWithCollisionMap(&g.asteroids[i])
	}
	for i := range g.lasers {
		g.registerWithCollisionMap(&g.lasers[i])
	}
}

func (g *Game) handleCollisions() {
	for i := range g.collisionMap.grid {
		for j := range g.collisionMap.grid[i] {
			if len(g.collisionMap.grid[i]) == 1 {
				break
			}
			for k := range g.collisionMap.grid[i] {
				if g.collisionMap.grid[i][j] == nil {
					logger.Panic("CollisionMap grid value at j is nil")
					break
				}
				if g.collisionMap.grid[i][k] == nil {
					logger.Panic("CollisionMap grid value at k is nil")
					break
				}

				if detectCollision(g.collisionMap.grid[i][j], g.collisionMap.grid[i][k]) {
					g.collisionMap.grid[i][j].HandleCollision(g.collisionMap.grid[i][k])
					g.collisionMap.grid[i][k].HandleCollision(g.collisionMap.grid[i][j])
				}
			}
		}
	}
}

func (g *Game) SpawnAsteroid() {
	if time.Since(lastEnemySpawn) > 1*time.Second {
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
		asteroidDividend := math.Abs(playerXDist) + math.Abs(playerYDist)/asteroidSpeed
		asteroidWidth := rand.Intn(40) + 10
		asteroidHeight := rand.Intn(40) + 10

		asteroid := Asteroid{
			x:      float64(xLocation),
			y:      float64(yLocation),
			dirX:   (playerXDist / asteroidDividend),
			dirY:   (playerYDist / asteroidDividend),
			sprite: ebiten.NewImage(asteroidWidth, asteroidHeight),
			width:  asteroidWidth,
			height: asteroidHeight,
		}
		asteroid.sprite.Fill(color.RGBA{177, 10, 75, 255})
		g.asteroids = append(g.asteroids, asteroid)
	}
}

func (g *Game) MoveAsteroids() {
	for i := range g.asteroids {
		g.asteroids[i].Update()
	}
}

func (a *Asteroid) Update() {
	a.x += a.dirX
	a.y += a.dirY
}

func (g *Game) MoveLasers() {
	for i := range g.lasers {
		g.lasers[i].Update()
	}
}

func (g *Game) LaserShoot() {
	currLaser := Laser{
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

func (g *Game) registerWithCollisionMap(c Collidable) {
	// width is 7
	b := c.GetBounds()
	minX := b.Min.X - b.Min.X%collisionGridSize                     //if 7, then start value is 5, cause round down
	minY := b.Min.Y - b.Min.Y%collisionGridSize                     //round down
	maxX := b.Max.X + collisionGridSize - b.Max.X%collisionGridSize // 14 round up to next 15 value 14 % 5 5-4
	maxY := b.Max.Y + collisionGridSize - b.Max.Y%collisionGridSize // round up
	for i := minX; i <= maxX; i += collisionGridSize {
		for j := minY; j <= maxY; j += collisionGridSize {
			key := fmt.Sprintf("%d:%d", i, j)
			g.collisionMap.grid[key] = append(g.collisionMap.grid[key], c)
		}
	}
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
	file, err := os.OpenFile("spacesurvival.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ebiten.SetWindowSize(1920, 1080)
	player := Player{
		height: 10,
		width:  10,
		dirX:   0,
		dirY:   1,
		x:      475,
		y:      265,
	}
	lasers := make([]Laser, 0, 500)
	asteroids := make([]Asteroid, 0, 25)
	player.createCharacter()
	player.character.Fill(color.White)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	cMap := CollisionMap{
		grid: make(map[string][]Collidable),
	}
	for i := screenWidth; i >= 0; i -= collisionGridSize {
		for j := screenHeight; j >= 0; j -= collisionGridSize {
			cMap.grid[fmt.Sprintf("%d:%d", i, j)] = make([]Collidable, 0, 100)
		}
	}
	//
	if err := ebiten.RunGame(&Game{player, lasers, asteroids, cMap}); err != nil {
		log.Fatal(err)
	}

}
