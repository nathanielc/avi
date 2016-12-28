package server

import (
	"bytes"
	"io"
	"sync"

	"github.com/golang/glog"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/gdvariant"
)

const scale = 0.01

type game struct {
	id          string
	frames      chan Frame
	clients     chan chan<- Frame
	deadClients chan chan<- Frame
	closing     chan struct{}

	sim    *avi.Simulation
	replay io.WriteCloser

	mu     sync.Mutex
	opened bool
	wg     sync.WaitGroup
}

func newGame(id string) *game {
	return &game{
		id:          id,
		frames:      make(chan Frame, 1),
		clients:     make(chan chan<- Frame, 10),
		deadClients: make(chan chan<- Frame, 10),
		closing:     make(chan struct{}),
	}
}

func (g *game) Open() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.opened {
		return nil
	}
	g.opened = true
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer g.replay.Close()
		g.Stream(g.replay)
	}()
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.sim.Start()
	}()
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.runDraw()
	}()
	return nil
}

func (g *game) Close() {
	close(g.closing)
	g.wg.Wait()
}

func (g *game) runDraw() {
	var clients []chan<- Frame

	var frames chan Frame

	for {
		select {
		case <-g.closing:
			for _, c := range clients {
				close(c)
			}
			return
		case client := <-g.deadClients:
			cs := clients[0:0]
			for _, c := range clients {
				if c != client {
					cs = append(cs, c)
				}
			}
			clients = cs
		case client := <-g.clients:
			clients = append(clients, client)
			// Block reading frames till we have the first client
			if frames == nil {
				frames = g.frames
			}
		case frame := <-frames:
			for _, c := range clients {
				c <- frame
			}
		}
	}
}

func (g *game) Draw(t float64, scores map[string]float64, new, existing []avi.Drawable, deleted []avi.ID) {
	var frame Frame
	frame.Time = float32(t)

	frame.Scores = make(map[string]float32, len(scores))
	for fleet, score := range scores {
		frame.Scores[fleet] = float32(score)
	}

	frame.NewObjects = make([]NewObject, len(new))
	for i, d := range new {
		p := d.Position()
		frame.NewObjects[i] = NewObject{
			ID: uint32(d.ID()),
			Position: gdvariant.Vector3{
				X: float32(p.X * scale),
				Y: float32(p.Y * scale),
				Z: float32(p.Z * scale),
			},
			Radius: float32(d.Radius() * scale),
			Model:  d.Texture(),
		}
	}

	frame.ObjectUpdates = make([]ObjectUpdate, len(existing))
	for i, d := range existing {
		p := d.Position()
		frame.ObjectUpdates[i] = ObjectUpdate{
			ID: uint32(d.ID()),
			Position: gdvariant.Vector3{
				X: float32(p.X * scale),
				Y: float32(p.Y * scale),
				Z: float32(p.Z * scale),
			},
		}
	}

	frame.DeletedObjects = make([]uint32, len(deleted))
	for i, v := range deleted {
		frame.DeletedObjects[i] = uint32(v)
	}
	g.frames <- frame
}

func (g *game) Stream(w io.Writer) {
	frames := make(chan Frame, 10)
	select {
	case <-g.closing:
		return
	case g.clients <- frames:
	}

	var buf bytes.Buffer
	enc := gdvariant.NewEncoder(&buf)
	for frame := range frames {
		if err := enc.Encode(frame); err != nil {
			glog.Errorln(err)
			continue
		}
		bytes := buf.Bytes()
		if err := gdvariant.WriteUint32(w, uint32(len(bytes))); err != nil {
			glog.Errorln(err)
			g.deadClients <- frames
			return
		}
		if _, err := w.Write(bytes); err != nil {
			glog.Errorln(err)
			g.deadClients <- frames
			return
		}
		buf.Reset()
	}
}
