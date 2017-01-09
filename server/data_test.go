package server

import (
	"io"

	"github.com/nathanielc/gdvariant"
)

//func TestReplaySeeker(t *testing.T) {
//	d, err := newData("testdata")
//	if err != nil {
//		t.Fatal(err)
//	}
//	r := d.NewReplay("test")
//	rs, err := r.Seeker()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	frameCount, size, fr, err := rs.Reader(0, 130*3)
//	if err != nil {
//		t.Fatal(err)
//	}
//	frames, err := readFrames(fr)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if got, exp := frameCount, 130; got != exp {
//		t.Errorf("unexpected frame count got %d exp %d", got, exp)
//	}
//	if got, exp := len(frames), frameCount; got != exp {
//		t.Errorf("unexpected number of decoded frames got %d exp %d", got, exp)
//	}
//	if got, exp := size, 451000; got != exp {
//		t.Errorf("unexpected frame size got %d exp %d", got, exp)
//	}
//	for i, f := range frames {
//		log.Println(i, f.Time)
//	}
//}

func readFrames(r io.Reader) ([]Frame, error) {
	var frames []Frame
	for {
		_, err := gdvariant.ReadInt32(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var frame Frame
		err = gdvariant.NewDecoder(r).Decode(&frame)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}
