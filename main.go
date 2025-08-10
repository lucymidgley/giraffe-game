package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	ScreenWidth  = 832 // 13 bricks wide
	ScreenHeight = 600
	BlockHeight  = 64
	BlockWidth   = 64
	GroundY      = ScreenHeight - 2*BlockHeight
	Gravity      = 0.1
)

//go:embed assets/*
var assets embed.FS

var Bricks = mustLoadImages("assets/bricks/*.png")

var ScoreFont = mustLoadFont("assets/font.ttf")

func mustLoadFont(name string) font.Face {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{Size: 48, DPI: 72, Hinting: font.HintingVertical})
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
}
type Brick struct {
	position Vector
	sprite   *ebiten.Image
}

func NewBrick(x, y float64) *Brick {
	sprite := BrickSprites[rand.Intn(len(Bricks))]

	pos := Vector{
		X: x,
		Y: ScreenHeight - BlockHeight - y,
	}
	return &Brick{
		position: pos,
		sprite:   sprite,
	}
}

func (g *Game) DrawGround() {
	for i := 0; i < 13; i++ {
		brick := NewBrick(float64(i*BlockWidth), 0)
		g.bricks = append(g.bricks, brick)
	}

	xCoords := []int{4, 5, 6, 7, 8, 9, 10, 11, 12}
	randomX := float64(xCoords[rand.Intn(len(xCoords))] * BlockWidth)
	brick1 := NewBrick(randomX, BlockHeight)
	brick2 := NewBrick(randomX, 2*BlockHeight)
	g.bricks = append(g.bricks, brick1, brick2)
}
func NewPlayer(game *Game) *Player {
	sprite := PlayerSprite

	pos := Vector{
		X: BlockWidth / 2,
		Y: GroundY,
	}
	return &Player{
		game:     game,
		position: pos,
		sprite:   sprite,
	}
}

func (p *Player) Update() {
	jumpingStrength := 0.5
	if p.isJumping {
		if p.position.Y >= GroundY {
			p.isJumping = false
			p.velocity = 0
		} else {
			p.velocity += Gravity
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.isJumping = true
		p.velocity -= jumpingStrength
	}
	p.position.Y += p.velocity
}

func (p *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	scaleX := 0.35
	scaleY := 0.35
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

func (b *Brick) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(b.position.X, b.position.Y)

	screen.DrawImage(b.sprite, op)
}

type Vector struct {
	X float64
	Y float64
}

type Game struct {
	score  int
	player *Player
	bricks []*Brick
}

func (g *Game) Update() error {
	g.player.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, b := range g.bricks {
		b.Draw(screen)
	}

	text.Draw(screen, fmt.Sprintf("%06d", g.score), ScoreFont, ScreenWidth/2-100, 50, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func NewGame() *Game {
	g := &Game{}
	g.player = NewPlayer(g)

	return g
}

func main() {
	g := NewGame()
	g.DrawGround()

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
