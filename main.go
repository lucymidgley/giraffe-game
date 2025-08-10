package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"

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

type Player struct {
	game     *Game
	position Vector
	sprite   *ebiten.Image
}

func NewPlayer(game *Game) *Player {
	sprite := PlayerSprite

	pos := Vector{
		X: BlockWidth / 2,
		Y: ScreenHeight - BlockWidth,
	}
	return &Player{
		game:     game,
		position: pos,
		sprite:   sprite,
	}
}

func (p *Player) Update() {
	speed := 5.0

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.position.Y -= speed
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	scaleX := 0.25
	scaleY := 0.25
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

type Vector struct {
	X float64
	Y float64
}

type Game struct {
	score  int
	player *Player
}

func (g *Game) Update() error {
	g.player.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)
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

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
