package main

import (
	"os"

	"gopkg.in/logex.v1"

	"bazil.org/fuse"
	"golang.org/x/net/context"
)

type Handler struct {
	file *os.File
}

func NewHandler(f *os.File) *Handler {
	h := &Handler{
		file: f,
	}
	return h
}

func (h *Handler) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	logex.Struct(*req)
	resp.Data = make([]byte, req.Size)
	n, err := h.file.ReadAt(resp.Data, req.Offset)
	resp.Data = resp.Data[:n]
	return err
}

func (h *Handler) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	logex.Struct(*req)
	n, err := h.file.WriteAt(req.Data, req.Offset)
	resp.Size = n
	return err
}

func (h *Handler) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	logex.Struct(*req)
	return h.file.Close()
}

func (h *Handler) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	logex.Struct(*req)
	return h.file.Sync()
}
