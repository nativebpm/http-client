package request

import "io"

type BodyType uint8

const (
	BodyTypeNone BodyType = iota
	BodyTypeBytes
	BodyTypeReader
	BodyTypeJSON
)

type ItemOp struct {
	Key, Value string
}

type BodyOp struct {
	Type          BodyType
	Data          []byte
	Reader        io.ReadCloser
	JSON          any
	ContentType   string
	ContentLength int64
}

type FileOp struct {
	Key      string
	Filename string
	Content  io.Reader
}
