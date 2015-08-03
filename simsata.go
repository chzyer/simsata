package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/chzyer/flagx"
	"gopkg.in/logex.v1"
)

var (
	FsName = "simsata"
)

type Config struct {
	Base   string `flag:"def=/Volumes/fuse"`
	Target string `flag:"def=/tmp/fuse"`
}

func NewConfig() *Config {
	var c Config
	flagx.Parse(&c)
	return &c
}

func main() {
	config := NewConfig()
	fuse.Unmount(config.Base)
	os.MkdirAll(config.Base, 0777)

	conn, err := process(config)
	if err != nil {
		logex.Fatal(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-c
	conn.Close()
	fuse.Unmount(config.Base)
}

func process(c *Config) (conn *fuse.Conn, err error) {
	if err = fuse.Unmount(c.Base); err == nil {
		logex.Info("last not unmount")
		time.Sleep(1000 * time.Millisecond)
		err = nil
	} else {
		err = nil
	}

	ops := []fuse.MountOption{
		fuse.AllowOther(),
		fuse.FSName(FsName),
		fuse.LocalVolume(),
		fuse.VolumeName("SimSATA"),
	}
	conn, err = fuse.Mount(c.Base, ops...)
	if err != nil {
		return nil, logex.Trace(err)
	}
	go fs.Serve(conn, NewTree("/", c.Target))
	logex.Info("connected.")
	return conn, nil
}

type Tree struct {
	root *Node
}

func NewTree(base, target string) *Tree {
	return &Tree{NewNode(base, target)}
}

func (t *Tree) Root() (fs.Node, error) {
	return t.root, nil
}
