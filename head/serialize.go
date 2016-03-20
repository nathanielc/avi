package head

//go:generate protoc --go_out=./ head.proto
import (
	"encoding/binary"
	"io"
	"log"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/protobuf/proto"
	"github.com/nathanielc/avi"
)

type protoStream struct {
	out       io.Writer
	varintBuf []byte
}

func NewProtoStream(out io.Writer, fps int) *protoStream {
	s := &protoStream{
		out:       out,
		varintBuf: make([]byte, binary.MaxVarintLen64),
	}
	// write header
	meta := Meta{
		FPS: int32(fps),
	}
	WriteMessage(out, &meta)
	return s
}

func (s *protoStream) Draw(scores map[string]float64, drawables []avi.Drawable, deleted []avi.ID) {
	var frame Frame
	frame.Objects = make([]*Object, len(drawables))
	frame.Scores = make([]*Score, len(scores))
	i := 0
	for fleet, score := range scores {
		frame.Scores[i] = &Score{
			Fleet: fleet,
			Score: score,
		}
		i++
	}
	for i, d := range drawables {
		p := d.Position()
		frame.Objects[i] = &Object{
			ID: uint64(d.ID()),
			Pos: &Vector{
				X: p.X(),
				Y: p.Y(),
				Z: p.Z(),
			},
			Radius:  d.Radius(),
			Texture: d.Texture(),
		}
	}

	frame.Deleted = make([]uint64, len(deleted))
	for i, v := range deleted {
		frame.Deleted[i] = uint64(v)
	}
	err := WriteMessage(s.out, &frame)
	if err != nil {
		log.Println(err)
		return
	}
}

func WriteMessage(w io.Writer, m proto.Message) error {
	var buf [binary.MaxVarintLen64]byte
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	n := binary.PutUvarint(buf[:], uint64(len(data)))
	_, err = w.Write(buf[:n])
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

type byteReader struct {
	io.Reader
	b []byte
}

func newByteReader(r io.Reader) *byteReader {
	return &byteReader{
		Reader: r,
		b:      make([]byte, 1),
	}
}

func (b *byteReader) ReadByte() (byte, error) {
	for {
		n, err := b.Reader.Read(b.b)
		if n == 1 {
			return b.b[0], err
		}
	}
}

type ByteReadReader interface {
	io.Reader
	io.ByteReader
}

func ReadMessage(readBuf *[]byte, r ByteReadReader, m proto.Message) error {
	s, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	size := int(s)
	if cap(*readBuf) < size {
		*readBuf = make([]byte, 0, size)
	}
	buf := (*readBuf)[:size]
	total := 0
	for total != size {
		n, err := r.Read(buf[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return proto.Unmarshal(buf[:total], m)
}

type proxyStream struct {
	streams []avi.Drawer
}

func NewProxyStream(s ...avi.Drawer) *proxyStream {
	return &proxyStream{
		streams: s,
	}
}

func (p proxyStream) Draw(scores map[string]float64, drawables []avi.Drawable, deleted []avi.ID) {
	for _, s := range p.streams {
		s.Draw(scores, drawables, deleted)
	}
}

func ProtoStreamUpdates(r io.Reader) <-chan FrameUpdate {
	updates := make(chan FrameUpdate, 1000)
	go func() {
		var readBuf []byte
		brr := newByteReader(r)
		// Read header
		var meta Meta
		err := ReadMessage(&readBuf, brr, &meta)
		if err != nil {
			panic(err)
		}
		delta := time.Second / time.Duration(meta.FPS)
		last := time.Now()
		for {
			var frame Frame
			err := ReadMessage(&readBuf, brr, &frame)
			if err == io.EOF {
				return
			}
			if err != nil {
				panic(err)
			}
			scores := make([]score, len(frame.Scores))
			for i, s := range frame.Scores {
				scores[i] = score{
					fleet: s.Fleet,
					score: s.Score,
				}
			}
			drawables := make([]avi.Drawable, len(frame.Objects))
			for i, o := range frame.Objects {
				drawables[i] = drawable{
					id: avi.ID(o.ID),
					position: mgl64.Vec3{
						o.Pos.X,
						o.Pos.Y,
						o.Pos.Z,
					},
					radius:  o.Radius,
					texture: o.Texture,
				}
			}
			deleted := make([]avi.ID, len(frame.Deleted))
			for i, d := range frame.Deleted {
				deleted[i] = avi.ID(d)
			}
			time.Sleep(delta - time.Since(last))
			last = time.Now()
			updates <- FrameUpdate{
				scores:    scores,
				drawables: drawables,
				deleted:   deleted,
			}
		}
	}()
	return updates
}

type drawable struct {
	id       avi.ID
	position mgl64.Vec3
	radius   float64
	texture  string
}

func (d drawable) ID() avi.ID {
	return d.id
}
func (d drawable) Position() mgl64.Vec3 {
	return d.position
}
func (d drawable) Radius() float64 {
	return d.radius
}
func (d drawable) Texture() string {
	return d.texture
}
