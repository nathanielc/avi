package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/gdvariant"
)

const scale = 0.01

type game struct {
	id             string
	replay         Replay
	replayStreamer *replayStreamer
	pos            int64

	sim      *avi.Simulation
	finished chan struct{}
	wc       io.WriteCloser
	enc      *gdvariant.Encoder
	buf      bytes.Buffer

	mu      sync.RWMutex
	running bool
	wg      sync.WaitGroup
}

func newGame(id string, replay Replay) (*game, error) {
	var replayStreamer *replayStreamer
	s, err := replay.Seeker()
	if err == nil {
		var err error
		replayStreamer, err = newReplayStreamerFromReader(s, replay)
		if err != nil {
			return nil, err
		}
	} else {
		replayStreamer = newReplayStreamer(replay)
	}
	return &game{
		id:             id,
		replay:         replay,
		replayStreamer: replayStreamer,
	}, nil
}

func (g *game) Start(sim *avi.Simulation) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.running {
		return errors.New("game already running, cannot start again")
	}
	g.running = true
	g.sim = sim
	g.finished = make(chan struct{})

	// Create write closer for the replay
	wc, err := g.replay.WriteCloser()
	if err != nil {
		return err
	}
	g.wc = wc
	g.enc = gdvariant.NewEncoder(&g.buf)

	// Encode metadata to buffer
	meta := Meta{
		FPS: float32(g.sim.CorrectedFPS()),
	}
	g.replayStreamer.SetMeta(meta)
	if err := g.encodeObj(meta); err != nil {
		return err
	}

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.sim.Start()
		g.wc.Close()
		close(g.finished)
		g.mu.Lock()
		g.running = false
		g.replay.Date = time.Now()
		g.mu.Unlock()
	}()
	return nil
}

func (g *game) Wait() {
	g.wg.Wait()
}

func (g *game) IsRunning() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.running
}

func (g *game) Info() Game {
	running := g.IsRunning()
	t := g.replay.Date
	if running {
		t = time.Now()
	}
	return Game{
		ID:     g.id,
		Date:   t,
		Active: running,
	}
}

func (g *game) Stream(startFrame, stopFrame int) (*Stream, error) {
	// Just return the available frames
	g.mu.RLock()
	defer g.mu.RUnlock()
	t := g.replayStreamer.TotalFrameCount()
	if startFrame >= t {
		startFrame = t - 1
	}
	if stopFrame >= t {
		stopFrame = t - 1
	}
	s, err := g.replayStreamer.Stream(startFrame, stopFrame)
	return s, err
}

func (g *game) Draw(t float64, scores map[string]float64, new, existing []avi.Drawable, deleted []avi.ID) {
	var frame Frame
	frame.Time = float32(t)

	frame.Scores = make(map[string]float32, len(scores))
	for fleet, score := range scores {
		frame.Scores[fleet] = float32(score)
	}

	frame.Objects = make([]Object, len(new)+len(existing))
	for i, d := range new {
		p := d.Position()
		frame.Objects[i] = Object{
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
	for i, d := range existing {
		p := d.Position()
		frame.Objects[i] = Object{
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

	//frame.ObjectUpdates = make([]ObjectUpdate, len(existing))
	//for i, d := range existing {
	//	p := d.Position()
	//	frame.ObjectUpdates[i] = ObjectUpdate{
	//		ID: uint32(d.ID()),
	//		Position: gdvariant.Vector3{
	//			X: float32(p.X * scale),
	//			Y: float32(p.Y * scale),
	//			Z: float32(p.Z * scale),
	//		},
	//	}
	//}

	frame.DeletedObjects = make([]uint32, len(deleted))
	for i, v := range deleted {
		frame.DeletedObjects[i] = uint32(v)
	}

	// Encode frame to buffer
	g.mu.Lock()
	g.replayStreamer.AddFramePos(g.pos)
	g.mu.Unlock()
	if err := g.encodeObj(frame); err != nil {
		glog.Infoln(err)
		return
	}
}

func (g *game) encodeObj(o interface{}) error {
	// Encode object to buffer
	if err := g.enc.Encode(o); err != nil {
		return err
	}
	defer g.buf.Reset()
	// Write encoded object to g.wc
	bytes := g.buf.Bytes()
	if err := gdvariant.WriteUint32(g.wc, uint32(len(bytes))); err != nil {
		return err
	}
	if _, err := g.wc.Write(bytes); err != nil {
		return err
	}
	g.pos += int64(len(bytes) + 4)
	return nil
}

type Replay struct {
	GameID string    `json:"game_id"`
	Date   time.Time `json:"date"`
	fpath  string
}

func (r Replay) WriteCloser() (io.WriteCloser, error) {
	f, err := os.Create(r.fpath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

func (r Replay) Seeker() (ReadSeekCloser, error) {
	f, err := os.Open(r.fpath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type replayStreamer struct {
	Meta   Meta
	frames []int64
	replay Replay
}

func newReplayStreamer(r Replay) *replayStreamer {
	return &replayStreamer{
		replay: r,
	}
}

func newReplayStreamerFromReader(r io.ReadSeeker, replay Replay) (*replayStreamer, error) {
	rs := &replayStreamer{
		replay: replay,
	}
	// Read/Seek all frames
	readMeta := false
	pos := int64(0)
	for {
		l, err := gdvariant.ReadInt32(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !readMeta {
			var meta Meta
			err := gdvariant.NewDecoder(r).Decode(&meta)
			if err != nil {
				return nil, err
			}
			rs.SetMeta(meta)
			readMeta = true
		} else {
			rs.AddFramePos(pos)
			_, err := r.Seek(int64(l), io.SeekCurrent)
			if err != nil {
				return nil, err
			}
		}
		pos += int64(l) + 4
	}
	return rs, nil
}

func (rs *replayStreamer) SetMeta(m Meta) {
	rs.Meta = m
}

func (rs *replayStreamer) AddFramePos(pos int64) {
	rs.frames = append(rs.frames, pos)
}

func (rs *replayStreamer) TotalFrameCount() int {
	return len(rs.frames)
}

func (rs *replayStreamer) Stream(startFrame, stopFrame int) (*Stream, error) {
	if stopFrame > len(rs.frames) {
		return nil, fmt.Errorf("invalid stop frame %d, total frames %d", stopFrame, len(rs.frames))
	}
	pos := rs.frames[startFrame]
	stopPos := rs.frames[stopFrame]
	r, err := rs.replay.Seeker()
	if err != nil {
		return nil, err
	}
	_, err = r.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return &Stream{
		Length:      stopPos - pos,
		FrameCount:  stopFrame - startFrame,
		StopFrame:   stopFrame,
		TotalFrames: len(rs.frames),
		FPS:         float64(rs.Meta.FPS),
		r:           r,
	}, nil
}

type Stream struct {
	Length      int64
	FrameCount  int
	StopFrame   int
	TotalFrames int
	FPS         float64
	r           io.ReadCloser
	pos         int64
}

func (s *Stream) Read(buf []byte) (int, error) {
	diff := int(s.Length - s.pos)
	if diff == 0 {
		return 0, io.EOF
	}
	if diff < len(buf) {
		buf = buf[:diff]
	}
	n, err := s.r.Read(buf)
	s.pos += int64(n)
	return n, err
}

func (s *Stream) Close() {
	s.r.Close()
}
