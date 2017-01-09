package main

import (
	"flag"
	"os"
	"os/user"
	"runtime/pprof"

	"github.com/golang/glog"
	"github.com/nathanielc/avi/server"
	_ "github.com/nathanielc/avi/ships"
)

var cpuProfile = flag.String("cpuprofile", "", "if defined save a cpu profile to path.")
var memProfile = flag.String("memprofile", "", "if defined save a mem profile to path.")
var bindAddr = flag.String("bind", "localhost:4242", "Network bind address")
var dataDir = flag.String("data", "data", "Data directory.")

func main() {

	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			glog.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			glog.Fatal(err)
		}
		defer f.Close()
		defer pprof.Lookup("heap").WriteTo(f, 0)
	}

	server, err := server.New(server.ServerConf{
		Addr:    *bindAddr,
		DataDir: *dataDir,
	})
	if err != nil {
		glog.Error(err)
		return
	}
	defer server.Close()
	err = server.Serve()
	if err != nil {
		glog.Error(err)
	}
}

func getHomeDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}
