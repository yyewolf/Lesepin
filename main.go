package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"time"

	_ "embed"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 1000
	screenHeight = 800
)

var (
	player          *ebiten.Image
	ennemies        *ebiten.Image
	spawnedEnnemies []*Ennemy
	lastEnnemy      time.Time
	user            = &User{Menu: true, Score: -5}
)

type User struct {
	Position   float64
	Score      int
	MenuSelect int
	Menu       bool
	HitScore   bool
	Pause      bool
	GameOver   bool
}

type Game struct {
	count int
	keys  []ebiten.Key

	// Relative to audio
	player       *audio.Player
	audioContext *audio.Context

	// Relative to inputs
	gamepadids []ebiten.GamepadID
}

const (
	sampleRate = 22255

	introLengthInSecond = 5
	loopLengthInSecond  = 4
)

func (g *Game) Update() error {
	/*
		Purely for debugging purpose
	*/
	for i := ebiten.StandardGamepadButton(0); i <= ebiten.StandardGamepadButtonMax; i++ {
		for _, id := range g.gamepadids {
			if inpututil.IsStandardGamepadButtonJustPressed(id, i) {
				fmt.Printf("Key %v is being pressed on Gamepad %v.\n", i, id)
			}
		}
	}
	/*
		----------------------------
	*/
	if !user.Menu {
		if user.GameOver {
			g.keyGameOverMenu()
		} else {
			if !user.Pause {
				g.count++
				g.moveEnnemies()
				g.movePlayer()
				g.pickEnnemy()
				if g.pressedEsc() {
					user.Pause = true
				}
			} else {
				g.keyPauseMenu()
			}
		}
	}
	g.gamepadids = inpututil.AppendJustConnectedGamepadIDs(g.gamepadids)
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	if g.player != nil {
		return nil
	}
	if g.audioContext == nil {
		g.audioContext = audio.NewContext(sampleRate)
	}
	if user.Menu {
		g.keyMenu()
		return nil
	}
	// Decode an Ogg file.
	// oggS is a decoded io.ReadCloser and io.Seeker.
	oggS, err := vorbis.Decode(g.audioContext, bytes.NewReader(audioLoop))
	if err != nil {
		return err
	}
	// Create an infinite loop stream from the decoded bytes.
	// s is still an io.ReadCloser and io.Seeker.
	s := audio.NewInfiniteLoopWithIntro(oggS, introLengthInSecond*4*sampleRate, loopLengthInSecond*4*sampleRate)
	g.player, err = g.audioContext.NewPlayer(s)
	if err != nil {
		return err
	}
	g.player.SetVolume(0.1)
	// Play the infinite-length stream. This never ends.
	g.player.Play()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if user.Menu {
		g.drawMenu(screen)
	} else {
		if user.GameOver {
			g.drawGameOverMenu(screen)
		} else {
			if user.Pause {
				g.drawPauseMenu(screen)
			} else {
				g.drawPlayer(screen)
				g.drawAllEnnemies(screen)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

//go:embed assets/player.png
var playerBytes []byte

//go:embed assets/enemy.png
var ennemyBytes []byte

//go:embed assets/loop.ogg
var audioLoop []byte

//go:embed assets/hit.ogg
var hitSound []byte

var (
	mplusNormalFont font.Face
	mplusBigFont    font.Face
)

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	mplusBigFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	img, _, err := image.Decode(bytes.NewReader(playerBytes))
	if err != nil {
		log.Fatal(err)
	}
	player = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(ennemyBytes))
	if err != nil {
		log.Fatal(err)
	}
	ennemies = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Le sépan")
	if err := ebiten.RunGame(&Game{gamepadids: []ebiten.GamepadID{}}); err != nil {
		log.Fatal(err)
	}
}
