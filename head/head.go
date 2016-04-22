package head

import (
	"fmt"
	_ "image/jpeg"
	"log"
	"path/filepath"
	"sort"

	"github.com/nathanielc/avi"

	"azul3d.org/engine/gfx"
	"azul3d.org/engine/gfx/gfxutil"
	"azul3d.org/engine/gfx/window"
	"azul3d.org/engine/lmath"
)

const minRadius = 3

var flatShader *gfx.Shader

func init() {
	// Load the flat shader.
	flatShader, _ = gfxutil.OpenShader("head/models/flat")
}

func Run(updates <-chan FrameUpdate) {
	go func() {
		// Create our window.
		props := window.NewProps()
		props.SetTitle("AVI {FPS}")
		props.SetPos(0, 0)
		props.SetSize(1280, 720)
		w, r, err := window.New(props)
		if err != nil {
			log.Fatal(err)
		}
		go gfxLoopWindow(w, r, updates)
	}()
	window.MainLoop()
}

func gfxLoopWindow(w window.Window, d gfx.Device, updates <-chan FrameUpdate) {
	// Create our events channel with sufficient buffer size.
	events := make(chan window.Event, 256)

	// Notify our channel anytime any event occurs.
	w.Notify(events, window.KeyboardTypedEvents)

	hud := newHud(d, d.Bounds())
	scoreDisplay := hud.addDisplay(lmath.Vec3{X: 1, Y: 1})

	models := make(map[string]*Model)
	objects := make(map[avi.ID]*gfx.Object)

	blueTexture, err := gfxutil.OpenTexture("head/models/blue.jpg")
	if err != nil {
		log.Fatal(err)
	}

	var event window.Event
	for {
		// Grab event if there was one
		select {
		case event = <-events:
		default:
			event = nil
		}

		// Wait for frame upate
		u := <-updates
		if cap(scoreDisplay.Lines) < len(u.scores) {
			scoreDisplay.Lines = make([]string, len(u.scores))
		} else {
			scoreDisplay.Lines = scoreDisplay.Lines[0:len(u.scores)]
		}
		for i, score := range u.scores {
			scoreDisplay.Lines[i] = fmt.Sprintf("%s: %0.1f", score.fleet, score.score)
		}
		sort.Strings(scoreDisplay.Lines)
		scoreDisplay.Changed = true
		hud.Update(event)
		for _, id := range u.deleted {
			if o := objects[id]; o != nil {
				o.Destroy()
			}
			delete(objects, id)
		}
		for _, o := range u.drawables {
			obj := objects[o.ID()]
			if obj == nil {
				obj = gfx.NewObject()
				obj.State = gfx.NewState()
				// Load model
				model, ok := models[o.Texture()]
				if !ok {
					models, err := openObjFile(filepath.Join("head/models", o.Texture()+".obj"))
					if err == nil && len(models) > 0 {
						model = models[0]
						model.Shader = flatShader
					}
				}
				// Default to sphere
				if model == nil {
					// Load the texture.
					texture, err := gfxutil.OpenTexture(filepath.Join("head/models", o.Texture()+".jpg"))
					if err != nil {
						texture = blueTexture
					}
					model = &Model{
						Meshes:   []*gfx.Mesh{unitSphereMesh},
						Textures: []*gfx.Texture{texture},
						Shader:   flatShader,
					}
				}
				models[o.Texture()] = model
				copy := model.Copy()
				obj.Meshes = copy.Meshes
				obj.Textures = copy.Textures
				obj.Shader = copy.Shader

				r := o.Radius()
				if r < minRadius {
					r = minRadius
				}
				obj.SetScale(lmath.Vec3{X: r, Y: r, Z: r})
				objects[o.ID()] = obj
			}
			pos := o.Position()
			obj.SetPos(lmath.Vec3{
				X: pos.X(),
				Y: pos.Y(),
				Z: pos.Z(),
			})
		}
		// Render the whole frame.
		d.Clear(d.Bounds(), gfx.Color{R: 0, G: 0, B: 0, A: 1})
		d.ClearDepth(d.Bounds(), 1.0)

		for _, o := range objects {
			d.Draw(d.Bounds(), o, hud.Camera)
		}

		hud.Draw()
		d.Render()
	}
}

type liveStream struct {
	upates chan FrameUpdate
}

func NewLiveStream() *liveStream {
	return &liveStream{
		upates: make(chan FrameUpdate),
	}
}

type score struct {
	fleet string
	score float64
}
type FrameUpdate struct {
	scores    []score
	drawables []avi.Drawable
	deleted   []avi.ID
}

func (s *liveStream) Draw(scores map[string]float64, drawables []avi.Drawable, deleted []avi.ID) {
	ss := make([]score, len(scores))
	i := 0
	for f, s := range scores {
		ss[i] = score{
			fleet: f,
			score: s,
		}
		i++
	}
	select {
	case s.upates <- FrameUpdate{
		scores:    ss,
		drawables: drawables,
		deleted:   deleted,
	}:
	default:
		//drop frame :(
	}
}

func (s *liveStream) Updates() <-chan FrameUpdate {
	return s.upates
}
