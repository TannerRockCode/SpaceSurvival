package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

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
var debugFrameRate bool
var profileApp bool

var lastEnemySpawn = time.Now()

type Game struct {
	score        int
	player       Player
	lasers       []Laser
	asteroids    []Asteroid
	crystals     []Crystal
	collisionMap CollisionMap
}

type CollisionMap struct {
	grid map[string][]Collidable
}

func (c *CollisionMap) Clear() {
	for i := range c.grid {
		clear(c.grid[i])
	}
}

func (g *Game) DrawLasers(screen *ebiten.Image) {
	for _, l := range g.lasers {
		l.sprite.Draw(screen)
	}
}

func (g *Game) DrawAsteroids(screen *ebiten.Image) {
	for _, a := range g.asteroids {
		a.sprite.Draw(screen)
	}
}

func (g *Game) DrawCrystals(screen *ebiten.Image) {
	for _, c := range g.crystals {
		c.sprite.Draw(screen)
	}
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
		} else {
			g.score += 1
		}
	}
	g.crystals = g.crystals[:i]
}

func (g *Game) MovePlayer() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.sprite.rad -= 0.1
		g.player.sprite.dirX = math.Sin(g.player.sprite.rad)
		g.player.sprite.dirY = math.Cos(g.player.sprite.rad)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.sprite.rad += 0.1
		g.player.sprite.dirX = math.Sin(g.player.sprite.rad)
		g.player.sprite.dirY = math.Cos(g.player.sprite.rad)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.sprite.y -= g.player.sprite.dirY
		g.player.sprite.x += g.player.sprite.dirX
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.sprite.y += g.player.sprite.dirY
		g.player.sprite.x -= g.player.sprite.dirX
	}
}

func (g *Game) PlayerShoot() {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		laser := Laser{
			sprite: Sprite{
				x:      g.player.sprite.x,
				y:      g.player.sprite.y,
				rad:    g.player.sprite.rad,
				dirX:   g.player.sprite.dirX * 1.3,
				dirY:   g.player.sprite.dirY * 1.3,
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
	GetBounds() image.Rectangle
	HandleCollision(Collidable)
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
					sprite: Sprite{
						x:      float64(a.sprite.x),
						y:      float64(a.sprite.y),
						dirX:   a.sprite.dirX*20 + randDirX,
						dirY:   a.sprite.dirY*20 + randDirY,
						rad:    a.sprite.rad,
						image:  ebiten.NewImage(cWidth, cHeight),
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

		playerXDist := (g.player.sprite.x - float64(xLocation))
		playerYDist := (g.player.sprite.y - float64(yLocation))
		asteroidDividend := math.Abs(playerXDist) + math.Abs(playerYDist)/asteroidSpeed
		asteroidWidth := rand.Intn(40) + 10
		asteroidHeight := rand.Intn(40) + 10

		asteroid := Asteroid{
			sprite: Sprite{
				x:      float64(xLocation),
				y:      float64(yLocation),
				dirX:   (playerXDist / asteroidDividend),
				dirY:   (playerYDist / asteroidDividend),
				rad:    0,
				image:  ebiten.NewImage(asteroidWidth, asteroidHeight),
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

func (g *Game) MoveAsteroids() {
	for i := range g.asteroids {
		g.asteroids[i].Update()
	}
}

func (g *Game) MoveLasers() {
	for i := range g.lasers {
		g.lasers[i].Update()
	}
}

func (g *Game) registerWithCollisionMap(c Collidable) {
	b := c.GetBounds()
	minX := b.Min.X - b.Min.X%collisionGridSize
	minY := b.Min.Y - b.Min.Y%collisionGridSize
	maxX := b.Max.X + collisionGridSize - b.Max.X%collisionGridSize
	maxY := b.Max.Y + collisionGridSize - b.Max.Y%collisionGridSize
	for i := minX; i <= maxX; i += collisionGridSize {
		for j := minY; j <= maxY; j += collisionGridSize {
			key := fmt.Sprintf("%d:%d", i, j)
			g.collisionMap.grid[key] = append(g.collisionMap.grid[key], c)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if debugFrameRate {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Ticks Per Second: %0.2f", ebiten.ActualTPS()))
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), 10, screenHeight-30)
	g.player.sprite.Draw(screen)
	g.DrawLasers(screen)
	g.DrawAsteroids(screen)
	g.DrawCrystals(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return screenWidth, screenHeight
}

func initDefaults() {
	debugFrameRate = false
	profileApp = false
	flag.BoolVar(&debugFrameRate, "dfr", false, "Enable debug mode")
	flag.BoolVar(&debugFrameRate, "debug-frame-rate", false, "Enable debug mode")
	flag.BoolVar(&profileApp, "p", false, "Enable profiling")
	flag.BoolVar(&profileApp, "profile", false, "Enable profiling")
	flag.Parse()
}

func main() {
	initDefaults()
	if profileApp {
		profile, err := os.Create("spacesurvival.prof")
		if err != nil {
			log.Fatal(err)
		}
		defer profile.Close()
		pprof.StartCPUProfile(profile)
		defer pprof.StopCPUProfile()
	}

	file, err := os.OpenFile("spacesurvival.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ebiten.SetWindowSize(1920, 1080)
	player := Player{
		sprite: Sprite{
			height: 10,
			width:  10,
			dirX:   0,
			dirY:   1,
			rad:    0,
			x:      475,
			y:      265,
		},
	}
	lasers := make([]Laser, 0, 500)
	asteroids := make([]Asteroid, 0, 50)
	crystals := make([]Crystal, 0, 250)
	player.createCharacter()
	player.sprite.image.Fill(color.White)
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
	if err := ebiten.RunGame(&Game{
		score:        0,
		player:       player,
		lasers:       lasers,
		asteroids:    asteroids,
		crystals:     crystals,
		collisionMap: cMap,
	}); err != nil {
		log.Fatal(err)
	}
}
