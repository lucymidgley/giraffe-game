package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
	"slices"

	"game/components"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	ScreenWidth     = 832 // 13 bricks wide
	ScreenHeight    = 600
	BrickHeight     = 64
	BrickWidth      = 64
	GroundY         = ScreenHeight - 2*BrickHeight
	Gravity         = 0.2
	ForwardMotion   = 2.0
	JumpingStrength = 6.0
	FrameWidth      = 13 * BrickWidth
)

//go:embed assets/*
var assets embed.FS

var Bricks = mustLoadImages("assets/bricks/*.png")

var ScoreFont = mustLoadFont("assets/font.ttf", 48.0)
var HighScoreFont = mustLoadFont("assets/font.ttf", 20.0)

func mustLoadFont(name string, size float64) font.Face {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingVertical})
	if err != nil {
		panic(err)
	}
	return face
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

var PlayerSprite = mustLoadImage("assets/giraffe.png")
var BrickSprites = mustLoadImages("assets/bricks/*.png")

type Player struct {
	game      *Game
	position  Vector
	sprite    *ebiten.Image
	velocity  float64
	isJumping bool
	jumpCount int
}
type Brick struct {
	position Vector
	sprite   *ebiten.Image
}

func NewBrick(x, y float64) *Brick {
	sprite := BrickSprites[rand.Intn(len(Bricks))]

	pos := Vector{
		X: x,
		Y: ScreenHeight - BrickHeight - y,
	}
	return &Brick{
		position: pos,
		sprite:   sprite,
	}
}

func (g *Game) DrawGround(x float64) {
	obstaclePostions := []int{rand.Intn(26), rand.Intn(26), rand.Intn(26)}
	for i := 0; i < 26; i++ {
		xCoord := float64(i*BrickWidth) + x
		brick := NewBrick(xCoord, 0)
		g.bricks = append(g.bricks, brick)
		if slices.Contains(obstaclePostions, i) {
			brick1 := NewBrick(xCoord, BrickHeight)
			brick2 := NewBrick(xCoord, 2*BrickHeight)
			g.bricks = append(g.bricks, brick1, brick2)
			obstacleRect := components.NewRect(brick1.position.X, brick1.position.Y, BrickWidth, 2*BrickHeight)
			g.obstacles = append(g.obstacles, &Obstacle{obstacleRect, false})
		}
	}
}
func NewPlayer(game *Game) *Player {
	sprite := PlayerSprite

	pos := Vector{
		X: BrickWidth / 2,
		Y: GroundY,
	}
	return &Player{
		game:     game,
		position: pos,
		sprite:   sprite,
	}
}

func (p *Player) Update() {
	if !p.game.freeze {
		p.position.X += ForwardMotion
	}
	if p.isJumping {
		if p.position.Y >= GroundY {
			p.isJumping = false
			p.velocity = 0
			p.jumpCount = 0
		} else {
			p.velocity += Gravity
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) && p.jumpCount < 3 {
		p.isJumping = true
		p.velocity -= JumpingStrength
		p.jumpCount += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.game.freeze = true
	}
	p.position.Y += p.velocity
}

func (p *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	scaleX := 0.35
	scaleY := 0.35
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(p.position.X, p.position.Y)
	op.GeoM.Translate(p.game.camera.X, p.game.camera.Y)

	screen.DrawImage(p.sprite, op)
}

func (p *Player) Collider() components.Rect {
	bounds := p.sprite.Bounds()
	return components.NewRect(
		p.position.X,
		p.position.Y,
		float64(bounds.Dx())*0.35,
		float64(bounds.Dy())*0.35,
	)
}

func (b *Brick) Draw(screen *ebiten.Image, g *Game) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(b.position.X, b.position.Y)
	op.GeoM.Translate(g.camera.X, g.camera.Y)

	screen.DrawImage(b.sprite, op)
}

type Vector struct {
	X float64
	Y float64
}
type Obstacle struct {
	bounds  components.Rect
	cleared bool
}

type Game struct {
	score            int
	player           *Player
	bricks           []*Brick
	obstacles        []*Obstacle
	freeze           bool
	camera           Vector
	nextDrawingPoint float64
	highScore        int
}

func (g *Game) Update() error {
	g.player.Update()
	if !g.freeze {
		for _, o := range g.obstacles {
			if o.bounds.Intersects(g.player.Collider()) {
				g.Reset()
			} else {
				if !o.cleared && g.player.position.X > o.bounds.MaxX() {
					o.cleared = true
					g.score++
					if g.highScore < g.score {
						g.highScore = g.score
					}
				}
			}
		}
	}
	if math.Abs(g.camera.X) >= g.nextDrawingPoint {
		g.DrawGround(g.nextDrawingPoint + FrameWidth)
		g.nextDrawingPoint = g.nextDrawingPoint + FrameWidth
	}
	g.camera.X -= ForwardMotion
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, b := range g.bricks {
		b.Draw(screen, g)
	}

	text.Draw(screen, fmt.Sprintf("%06d", g.score), ScoreFont, 30, 50, color.White)
	text.Draw(screen, fmt.Sprintf("High Score: "+"%06d", g.highScore), HighScoreFont, ScreenWidth-270, 30, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func NewGame() *Game {
	g := &Game{nextDrawingPoint: FrameWidth}
	g.player = NewPlayer(g)

	return g
}

func (g *Game) Reset() {
	g.player = NewPlayer(g)
	g.bricks = nil
	g.score = 0
	g.obstacles = nil
	g.DrawGround(0)
	g.camera = Vector{}
	g.nextDrawingPoint = FrameWidth
}

func main() {
	g := NewGame()
	g.DrawGround(0)

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
