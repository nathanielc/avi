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
	finished    chan struct{}
	closing     chan struct{}

	sim       *avi.Simulation
	replay    Replay
	wc        io.WriteCloser
	enc       *gdvariant.Encoder
	buf       bytes.Buffer
	writeCond *sync.Cond

	mu     sync.RWMutex
	opened bool
	wg     sync.WaitGroup
}

func newGame(id string, replay Replay, finished chan struct{}) *game {
	return &game{
		id:          id,
		replay:      replay,
		finished:    finished,
		frames:      make(chan Frame, 1),
		clients:     make(chan chan<- Frame, 10),
		deadClients: make(chan chan<- Frame, 10),
		closing:     make(chan struct{}),
		writeCond:   sync.NewCond(new(sync.Mutex)),
	}
}

func (g *game) Open() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.opened {
		return nil
	}
	g.opened = true
	wc, err := g.replay.WriteCloser()
	if err != nil {
		return err
	}
	g.wc = wc
	g.enc = gdvariant.NewEncoder(&g.buf)

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.sim.Start()
		close(g.finished)
		g.writeCond.Broadcast()
	}()
	return nil
}

func (g *game) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.opened {
		return
	}
	g.opened = false
	g.wc.Close()
	close(g.closing)
	g.wg.Wait()
}

func (g *game) Stream(w io.Writer) {
	rc, err := g.replay.ReadCloser()
	if err != nil {
		glog.Errorln(err)
		return
	}
	defer rc.Close()
	// Copy file until we are finished
	buf := make([]byte, 32*1024)
	for {
		if _, err := io.CopyBuffer(w, rc, buf); err != nil && err != io.EOF {
			glog.Errorln(err)
			return
		}
		select {
		case <-g.finished:
			return
		default:
			// Wait for next write
			g.writeCond.L.Lock()
			g.writeCond.Wait()
			g.writeCond.L.Unlock()
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

	// Encode frame to buffer
	if err := g.enc.Encode(frame); err != nil {
		glog.Errorln(err)
		return
	}
	defer g.buf.Reset()
	// Write encoded frame to g.wc
	bytes := g.buf.Bytes()
	if err := gdvariant.WriteUint32(g.wc, uint32(len(bytes))); err != nil {
		glog.Errorln(err)
		return
	}
	if _, err := g.wc.Write(bytes); err != nil {
		glog.Errorln(err)
		return
	}
	g.writeCond.Broadcast()
}
