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

	//"runtime/pprof"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth         = 960
	screenHeight        = 540
	asteroidSpeed       = 0.9
	crystalAcceleration = 5
	collisionGridSize   = 50
)

var logger *log.Logger

var lastEnemySpawn = time.Now()

type Game struct {
	player       Player
	lasers       []Laser
	asteroids    []Asteroid
	crystals     []Crystal
	collisionMap CollisionMap
}

type CollisionMap struct {
	grid map[string][]Collidable
}

type Sprite struct {
	x      float64
	y      float64
	rad    float64
	dirX   float64
	dirY   float64
	image   *ebiten.Image
	width  int
	height int
}

type Laser struct {
	sprite Sprite
	usedUp bool
}

type Asteroid struct {
	sprite Sprite
	destroyed bool
}

type Crystal struct {
	sprite Sprite
	absorbed bool
}

func (c *CollisionMap) Clear() {
	for i := range c.grid {
		clear(c.grid[i])
	}
}

func (l *Laser) createBeam() {
	l.sprite.image = ebiten.NewImage(l.sprite.width, l.sprite.height)
	l.sprite.image.Fill(color.RGBA{255, 0, 0, 255})
}

func (l *Laser) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(-float64(l.sprite.width)/2, -float64(l.sprite.height)/2)
	geo.Rotate(l.sprite.rad)
	geo.Translate(l.sprite.x, l.sprite.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(l.sprite.image, op)
}

func (g *Game) DrawLasers(screen *ebiten.Image) {
	for _, l := range g.lasers {
		l.Draw(screen)
	}
}

func (a *Asteroid) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(a.sprite.x, a.sprite.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(a.sprite.image, op)
}

func (g *Game) DrawAsteroids(screen *ebiten.Image) {
	for _, a := range g.asteroids {
		a.Draw(screen)
	}
}

func (c *Crystal) Draw(screen *ebiten.Image) {
	geo := ebiten.GeoM{}
	geo.Translate(c.sprite.x, c.sprite.y)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	screen.DrawImage(c.sprite.image, op)
}

func (g *Game) DrawCrystals(screen *ebiten.Image) {
	for _, c := range g.crystals {
		c.Draw(screen)
	}
}

func (l Laser) IsOffScreen() bool {
	return l.sprite.x < 0 || l.sprite.x > float64(screenWidth) || l.sprite.y < 0 || l.sprite.y > float64(screenHeight)
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
	return a.sprite.x < 0 || a.sprite.x > float64(screenWidth) || a.sprite.y < 0 || a.sprite.y > float64(screenHeight)
}

func (g *Game) CleanAsteroids() {
	i := 0
	for _, a := range g.asteroids {
		if !a.IsOffScreen() && !a.destroyed {
			g.asteroids[i] = a
			i++
		}
	}
	g.asteroids = g.asteroids[:i]
}

func (g *Game) CleanCrystals() {
 i := 0
 for _, c := range g.crystals {
	if !c.absorbed {
		g.crystals[i] = c
		i++
	}
 }
 g.crystals = g.crystals[:i]
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
			sprite : Sprite {
				x:      g.player.x,
				y:      g.player.y,
				rad:    g.player.rad,
				dirX:   g.player.dirX * 1.3,
				dirY:   g.player.dirY * 1.3,
				width:  1,
				height: 3,
			},
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

func (c Crystal) GetBounds() image.Rectangle {
	return image.Rect(int(c.sprite.x), int(c.sprite.y), int(c.sprite.x+float64(c.sprite.width)), int(c.sprite.y+float64(c.sprite.height)))
}

func (p Player) GetBounds() image.Rectangle {
	// Create a GeoM transformation matrix
	geo := ebiten.GeoM{}
	geo.Translate(-float64(p.width)/2, -float64(p.height)/2)
	geo.Rotate(p.rad)
	geo.Translate(p.x, p.y)

	// Get the transformed corners of the player
	x0, y0 := geo.Apply(0, 0)
	x1, y1 := geo.Apply(float64(p.width), 0)
	x2, y2 := geo.Apply(0, float64(p.height))
	x3, y3 := geo.Apply(float64(p.width), float64(p.height))

	// Find the bounding box that contains all the transformed points
	minX := math.Min(math.Min(x0, x1), math.Min(x2, x3))
	maxX := math.Max(math.Max(x0, x1), math.Max(x2, x3))
	minY := math.Min(math.Min(y0, y1), math.Min(y2, y3))
	maxY := math.Max(math.Max(y0, y1), math.Max(y2, y3))

	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

func (l Laser) GetBounds() image.Rectangle {
	// Create a GeoM transformation matrix
	geo := ebiten.GeoM{}
	geo.Translate(-float64(l.sprite.width)/2, -float64(l.sprite.height)/2)
	geo.Rotate(l.sprite.rad)
	geo.Translate(l.sprite.x, l.sprite.y)

	// Get the transformed corners of the laser
	x0, y0 := geo.Apply(0, 0)
	x1, y1 := geo.Apply(float64(l.sprite.width), 0)
	x2, y2 := geo.Apply(0, float64(l.sprite.height))
	x3, y3 := geo.Apply(float64(l.sprite.width), float64(l.sprite.height))

	// Find the bounding box that contains all the transformed points
	minX := math.Min(math.Min(x0, x1), math.Min(x2, x3))
	maxX := math.Max(math.Max(x0, x1), math.Max(x2, x3))
	minY := math.Min(math.Min(y0, y1), math.Min(y2, y3))
	maxY := math.Max(math.Max(y0, y1), math.Max(y2, y3))

	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

func (a Asteroid) GetBounds() image.Rectangle {
	return image.Rect(int(a.sprite.x), int(a.sprite.y), int(a.sprite.x+float64(a.sprite.width)), int(a.sprite.y+float64(a.sprite.height)))
}

func detectCollision(c1, c2 Collidable) bool {
	_, ok := c1.(*Laser)
	if ok {
		_, ok := c2.(*Laser)
		if ok {
			return false
		}
	}
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
	l, ok := c.(*Laser)
	if ok {
		a.sprite.dirX = l.sprite.dirX
		a.sprite.dirY = l.sprite.dirY
		a.destroyed = true
	}
}

//for each value in key, for each other value in key if(detectCollision(key1, key2) then key1.handleCollision(key2) key2.handleCollision(key1)

// hashmap grid 96 columns 54 rows - key, which is a combo of the column and row - value is a slice of collidable objects
// register. For every object in a key, detect collision.
//for every value in a key, check to see if that value is colliding with the other values
//c.detectCollision(c2)

func (g *Game) Update() error {
	g.clearCollisionMap()
	g.MoveLasers()
	g.MoveAsteroids()
	g.MoveCrystals()
	g.registerCollidables()
	g.handleCollisions()
	g.CreateCrystals()
	g.MovePlayer()
	g.PlayerShoot()
	g.SpawnAsteroid()
	g.CleanLasers()
	g.CleanAsteroids()
	g.CleanCrystals()

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
		if len(g.collisionMap.grid[i]) <= 1 {
			continue
		}
		collisionSlice := g.collisionMap.grid[i]
		for j := 0; j < len(collisionSlice); j++ {
			for k := j + 1; k < len(collisionSlice); k++ {
				if detectCollision(collisionSlice[j], collisionSlice[k]) {
					collisionSlice[j].HandleCollision(collisionSlice[k])
					collisionSlice[k].HandleCollision(collisionSlice[j])
				}
			}
		}
	}
}

func (g *Game) CreateCrystals() {
	for _, a := range g.asteroids {
		if a.destroyed {
			numCrystals := a.sprite.width * a.sprite.height / 100
			for i := 0; i < numCrystals; i++ {
				cHeight := 8 + rand.Intn(4)
				cWidth := 8 + rand.Intn(4)
				randDirX := rand.Float64() * 10
				randDirY := rand.Float64() * 10
				negPosX := rand.Intn(2)
				negPosY := rand.Intn(2)
				if negPosX == 0 {
					randDirX *= -1
				}
				if negPosY == 0 {
					randDirY *= -1
				}

				crystal := Crystal{
					sprite : Sprite {
						x:      float64(a.sprite.x),
						y:      float64(a.sprite.y),
						dirX:   a.sprite.dirX*20 + randDirX,
						dirY:   a.sprite.dirY*20 + randDirY,
						image: ebiten.NewImage(cWidth, cHeight),
						width:  cWidth,
						height: cHeight,
					},
				}
				crystal.sprite.image.Fill(color.RGBA{209, 60, 219, 200})
				g.crystals = append(g.crystals, crystal)
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
			sprite : Sprite{
				x:      float64(xLocation),
				y:      float64(yLocation),
				dirX:   (playerXDist / asteroidDividend),
				dirY:   (playerYDist / asteroidDividend),
				image: ebiten.NewImage(asteroidWidth, asteroidHeight),
				width:  asteroidWidth,
				height: asteroidHeight,
			},
		}
		asteroid.sprite.image.Fill(color.RGBA{177, 10, 75, 255})
		g.asteroids = append(g.asteroids, asteroid)
	}
}

func (g *Game) MoveCrystals() {
	for i := range g.crystals {
		g.crystals[i].Update(g.player)
	}
}

func (c *Crystal) Update(p Player) {
	playerXDist := (p.x - float64(c.sprite.x))
	playerYDist := (p.y - float64(c.sprite.y))
	crystalDividend := math.Abs(playerXDist) + math.Abs(playerYDist)/crystalAcceleration
	forceX := playerXDist / crystalDividend
	forceY := playerYDist / crystalDividend
	c.sprite.dirX = c.sprite.dirX*.9 + forceX
	c.sprite.dirY = c.sprite.dirY*.9 + forceY
	c.sprite.x += c.sprite.dirX
	c.sprite.y += c.sprite.dirY

	if c.GetBounds().Overlaps(p.GetBounds()) {
		c.absorbed = true;
	}
}

func (g *Game) MoveAsteroids() {
	for i := range g.asteroids {
		g.asteroids[i].Update()
	}
}

func (a *Asteroid) Update() {
	a.sprite.x += a.sprite.dirX
	a.sprite.y += a.sprite.dirY
}

func (g *Game) MoveLasers() {
	for i := range g.lasers {
		g.lasers[i].Update()
	}
}

func (l *Laser) Update() {
	l.sprite.x += l.sprite.dirX
	l.sprite.y -= l.sprite.dirY
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
	g.DrawCrystals(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return screenWidth, screenHeight
}

func main() {
	// profile, err := os.Create("spacesurvival.prof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer profile.Close()
	// pprof.StartCPUProfile(profile)
	// defer pprof.StopCPUProfile()

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
	asteroids := make([]Asteroid, 0, 50)
	crystals := make([]Crystal, 0, 250)
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
	if err := ebiten.RunGame(&Game{player, lasers, asteroids, crystals, cMap}); err != nil {
		log.Fatal(err)
	}

}
