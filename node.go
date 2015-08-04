package main

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/logex.v1"
)

type Node struct {
	Pwd    string
	Target string
}

func NewNode(pwd string, target string) *Node {
	return &Node{pwd, target}
}

func (t *Node) lookup(name string) fs.Node {
	if _, err := os.Stat(t.getTargetPath(name)); err != nil {
		return nil
	}
	return NewNode(t.getBasePath(name), t.getTargetPath(name))
}

func (t *Node) Attr(ctx context.Context, a *fuse.Attr) error {
	fi, err := os.Stat(t.getTargetPath(""))
	if err != nil {
		return err
	}
	st := fi.Sys().(*syscall.Stat_t)
	a.Nlink = uint32(st.Nlink)
	a.Size = uint64(st.Size)
	a.Gid = st.Gid
	a.Uid = st.Uid
	a.Ctime = time.Unix(st.Ctimespec.Unix())
	a.Blocks = uint64(st.Blocks)
	a.BlockSize = uint32(st.Blksize)
	a.Inode = st.Ino
	a.Mode = fi.Mode()
	logex.Struct(a)
	return nil
}

func (t *Node) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logex.Info("lookup", name)
	n := t.lookup(name)
	if n != nil {
		return n, nil
	}
	return nil, fuse.ENOENT
}

func (t *Node) getTargetPath(name string) string {
	return filepath.Join(t.Target, name)
}

func (t *Node) getBasePath(name string) string {
	return filepath.Join(t.Pwd, name)
}

func (t *Node) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	logex.Struct(req)
	if err := os.Mkdir(filepath.Join(t.Target, req.Name), req.Mode); err != nil {
		logex.Error(err)
		return nil, err
	}
	return NewNode(t.getBasePath(req.Name), t.getTargetPath(req.Name)), nil
}

func (t *Node) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return readDirAll(t.getTargetPath(""))
}

func readDirAll(targetPath string) ([]fuse.Dirent, error) {
	f, err := os.Open(targetPath)
	if err != nil {
		return nil, err
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	out := make([]fuse.Dirent, 0, len(fis))
	for _, fi := range fis {
		f := fis[0].Sys().(*syscall.Stat_t)
		out = append(out, fuse.Dirent{
			Name:  fi.Name(),
			Type:  fuse.DT_File,
			Inode: f.Ino,
		})
	}

	logex.Struct(out)
	return out, nil
}

func (t *Node) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	logex.Struct(t.Pwd, *req)

	target := t.getTargetPath("")
	f, err := os.OpenFile(target, int(req.Flags), 0777)
	if err != nil {
		logex.Error(err)
		return nil, err
	}
	return NewHandler(target, f), nil
}

func (t *Node) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	logex.Info(t.Pwd)
	target := t.getTargetPath(req.Name)
	base := t.getBasePath(req.Name)
	f, err := os.OpenFile(target, int(req.Flags), req.Mode)
	if err != nil {
		logex.Error(err)
		return nil, nil, err
	}

	node := NewNode(base, target)
	handler := NewHandler(target, f)
	return node, handler, nil
}

func (t *Node) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	logex.Struct(req)
	os.Remove(t.getTargetPath(req.Name))
	return nil
}

func (t *Node) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	logex.Struct(req, ctx)
	return nil
}

func (t *Node) Mknod(ctx context.Context, req *fuse.MknodRequest) (fs.Node, error) {
	logex.Struct(req)
	return nil, io.EOF
}

func (t *Node) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	logex.Struct(*req, ctx)
	return nil
}
