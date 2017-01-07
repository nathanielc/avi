package server_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/nathanielc/avi/server"
	"github.com/nathanielc/gdvariant"
)

func TestObject(t *testing.T) {
	testCases := []struct {
		obj server.Object
	}{
		{
			obj: server.Object{
				ID:       42,
				Position: gdvariant.Vector3{X: 1, Y: 2, Z: 3},
				Radius:   6,
				Model:    "borg",
			},
		},
		{
			obj: server.Object{
				ID:       4,
				Position: gdvariant.Vector3{X: -1, Y: -2, Z: -3},
				Radius:   -66,
				Model:    "borg",
			},
		},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		if err := gdvariant.NewEncoder(&buf).Encode(tc.obj); err != nil {
			t.Fatal(err)
		}
		var got server.Object
		if err := gdvariant.NewDecoder(&buf).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.obj) {
			t.Errorf("unexpected object:\ngot\n%+v\nexp\n%+v\n", got, tc.obj)
		}
	}
}

//func TestFrame(t *testing.T) {
//	testCases := []struct {
//		frame server.Frame
//	}{
//		{
//			frame: server.Frame{
//				Scores: map[string]float32{
//					"orange": -6,
//					"blue":   100,
//					"green":  67,
//				},
//				Objects: []server.Object{
//					{
//						ID:       42,
//						Position: gdvariant.Vector3{X: 1, Y: 2, Z: 3},
//						Radius:   6,
//						Model:    "borg",
//					},
//					{
//						ID:       2,
//						Position: gdvariant.Vector3{X: -1, Y: -2, Z: -3},
//						Radius:   -66,
//						Model:    "borg",
//					},
//				},
//				ObjectUpdates: []server.ObjectUpdate{
//					{
//						ID:       52,
//						Position: gdvariant.Vector3{X: 1, Y: 2, Z: 3},
//					},
//					{
//						ID:       22,
//						Position: gdvariant.Vector3{X: -1, Y: -2, Z: -3},
//					},
//				},
//				DeletedObjects: []uint32{12, 15},
//			},
//		},
//	}
//	for _, tc := range testCases {
//		var buf bytes.Buffer
//		if err := gdvariant.NewEncoder(&buf).Encode(tc.frame); err != nil {
//			t.Fatal(err)
//		}
//		var got server.Frame
//		if err := gdvariant.NewDecoder(&buf).Decode(&got); err != nil {
//			t.Fatal(err)
//		}
//		if !reflect.DeepEqual(got, tc.frame) {
//			t.Errorf("unexpected frame:\ngot\n%+v\nexp\n%+v\n", got, tc.frame)
//		}
//	}
//}
