package head

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math"

	"azul3d.org/engine/gfx"
	"azul3d.org/engine/gfx/camera"
	"azul3d.org/engine/gfx/window"
	"azul3d.org/engine/keyboard"
	"azul3d.org/engine/lmath"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var hudDistance = lmath.Vec3{Y: 1}

var courierNewFont *truetype.Font

func init() {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile("head/models/cour.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	courierNewFont = f
}

func newHud(d gfx.Device, b image.Rectangle) *hud {
	h := &hud{
		device: d,
	}
	cam := camera.New(b)
	cam.Far = 10000
	cam.Update(b)
	h.Camera = cam

	return h
}

type hud struct {
	device   gfx.Device
	displays []*display

	Camera *camera.Camera
}

func (h *hud) addDisplay(pos lmath.Vec3) *display {
	// Create a card mesh.
	cardMesh := gfx.NewMesh()
	cardMesh.Vertices = []gfx.Vec3{
		// Bottom-left triangle.
		{-1, 0, -1},
		{1, 0, -1},
		{-1, 0, 1},

		// Top-right triangle.
		{-1, 0, 1},
		{1, 0, -1},
		{1, 0, 1},
	}
	cardMesh.TexCoords = []gfx.TexCoordSet{{Slice: []gfx.TexCoord{
		{0, 1},
		{1, 1},
		{0, 0},

		{0, 0},
		{1, 1},
		{1, 0},
	}}}

	// Create a card object.
	card := gfx.NewObject()
	card.State = gfx.NewState()
	card.AlphaMode = gfx.AlphaToCoverage
	card.Shader = flatShader
	tex := gfx.NewTexture()
	card.Textures = []*gfx.Texture{tex}
	card.Meshes = []*gfx.Mesh{cardMesh}
	card.SetPos(h.Camera.Pos().Add(pos).Add(hudDistance))
	d := &display{
		card:    card,
		texture: tex,
	}
	d.Update()
	h.displays = append(h.displays, d)
	return d
}

func (h *hud) Draw() {
	for _, display := range h.displays {
		h.device.Draw(h.device.Bounds(), display.card, h.Camera)
	}
}

func (h *hud) Update(event window.Event) {
	if event != nil {
		switch e := event.(type) {
		case keyboard.Typed:
			switch e.S {
			case "a":
				h.Translate(lmath.Vec3{X: -1})
			case "d":
				h.Translate(lmath.Vec3{X: 1})
			case "w":
				h.Translate(lmath.Vec3{Y: 1})
			case "s":
				h.Translate(lmath.Vec3{Y: -1})
			case "q":
				h.Translate(lmath.Vec3{Z: 1})
			case "e":
				h.Translate(lmath.Vec3{Z: -1})
			}
		}
	}
	for _, display := range h.displays {
		if display.Changed {
			display.Update()
		}
	}
}

func (h *hud) Translate(vec lmath.Vec3) {
	h.Camera.SetPos(h.Camera.Pos().Add(vec))
	for _, display := range h.displays {
		display.card.SetPos(display.card.Pos().Add(vec))
	}
}

type display struct {
	card    *gfx.Object
	texture *gfx.Texture
	Lines   []string
	Changed bool
}

func (d *display) Update() {
	const imgW, imgH = 640, 480
	const size, dpi, spacing = 13, 256, 1.5
	fg, bg := image.NewUniform(color.NRGBA{R: 6, G: 242, A: 255}), image.Transparent
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	// Draw the text.
	drawer := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(courierNewFont, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}
	y := 10 + int(math.Ceil(size*dpi/72))
	dy := int(math.Ceil(size * spacing * dpi / 72))
	for _, s := range d.Lines {
		drawer.Dot = fixed.P(10, y)
		drawer.DrawString(s)
		y += dy
	}

	d.texture.Loaded = false
	d.texture.Source = rgba
	d.texture.Bounds = rgba.Bounds()
}
