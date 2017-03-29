//Map_Gen Version 0.01
//
//Created: 			3/28/2017
//Last modified:	3/28/2017

//Notes: Newlines in the strings may only be compatible with a window enviorment.

package main


	
import (
    "fmt"
    "log"
	"math/rand"
    "os"
	"time"
)

func main() {
    map_file, err := os.Create("map_gen_v0.01.txt")
    if err != nil {
        log.Fatal("Cannot create file", err)
    }
    defer map_file.Close()

    initial_text  := "---\r\nradius: 1e6\r\n\r\nrules:\r\n  score: 500\r\n  max_fleet_mass: 1e5\r\nstarting_points:\r\n  - [1000,0,0]\r\n  - [0,1000,0]\r\ncontrol_points:\r\n  - mass: 1e6\r\n    radius: 1e1\r\n    position: [0, 0 , 0]\r\n    points: 1\r\n    influence: 140\r\nasteroids:\r\n  - mass: 1e6\r\n    radius: 6e1\r\n    position: "
	fmt.Fprintf(map_file, initial_text)

	rand.Seed(time.Now().UTC().UnixNano())
	asteroids_position := []string{
		"[100, 0, 0]\r\n",
		"[-100, 0, 0]\r\n",
		"[0, 100, 0]\r\n",
		"[0, -100, 0]\r\n",
		"[0, 0, 100]\r\n",
		"[0, 0, -100]\r\n",
	}
	fmt.Fprintf(map_file, asteroids_position[rand.Intn(len(asteroids_position))])
}
