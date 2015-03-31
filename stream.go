package avi

import (
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/nvcook42/avi/head"
	"io"
)

var strm = &head.Stream{
	Frame: make([]*head.Frame, 1, 1),
}

type Stream struct {
	out io.Writer
}

func NewStream(out io.Writer) *Stream {
	return &Stream{
		out: out,
	}
}

func (s *Stream) SendFrame(ships []*shipT, projs []*projectile, astds []*asteroid, ctlps []*controlPoint) {
	frame := &head.Frame{
		Object: make([]*head.Object, 0, len(ships)+len(projs)),
	}
	for _, ship := range ships {
		appendObject(frame, ship, head.Texture_SHIP, ship.texture)
	}
	for _, ctlp := range ctlps {
		appendObject(frame, ctlp, head.Texture_CONTROL_POINT, "")
	}
	for _, astd := range astds {
		appendObject(frame, astd, head.Texture_ASTEROID, "")
	}
	for _, proj := range projs {
		appendObject(frame, proj, head.Texture_PROJECTILE, "")
	}

	strm.Frame[0] = frame

	data, err := proto.Marshal(strm)
	if err != nil {
		glog.Errorln(err)
		return
	}
	s.out.Write(data)
}

func appendObject(frame *head.Frame, obj Object, tex head.Texture, customTexture string) {
	p := obj.GetPosition()
	x := float32(p.X())
	y := float32(p.Y())
	z := float32(p.Z())
	pos := &head.Vector{
		X: &x,
		Y: &y,
		Z: &z,
	}

	r := float32(obj.GetRadius())

	object := &head.Object{
		ID:     proto.Int64(obj.GetID()),
		Pos:    pos,
		Radius: &r,
		Tex:    &tex,
	}

	if len(customTexture) > 0 {
		object.TexCustom = proto.String(customTexture)
	}

	frame.Object = append(frame.Object, object)
}
