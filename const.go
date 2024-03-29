package lsgo

import (
	"errors"
	"fmt"
)

type FileVersion uint32

const (
	// Initial version of the LSF format
	VerInitial FileVersion = iota + 1

	// LSF version that added chunked compression for substreams
	VerChunkedCompress

	// LSF version that extended the node descriptors
	VerExtendedNodes

	// BG3 version, no changes found so far apart from version numbering
	VerBG3

	// Latest version supported by this library
	MaxVersion = iota
)

type CompressionMethod int

const (
	CMInvalid CompressionMethod = iota - 1
	CMNone
	CMZlib
	CMLZ4
)

type CompressionLevel int

const (
	FastCompression    CompressionLevel = 0x10
	DefaultCompression CompressionLevel = 0x20
	MaxCompression     CompressionLevel = 0x40
)

var (
	ErrVectorTooBig    = errors.New("the vector is too big cannot marshal to an xml element")
	ErrInvalidNameKey  = errors.New("invalid name key")
	ErrKeyDoesNotMatch = errors.New("key for this node does not match")
)

type HeaderError struct {
	Expected string
	Got      []byte
}

func (he HeaderError) Error() string {
	return fmt.Sprintf("Invalid LSF signature; expected % X, got % X", he.Expected, he.Got)
}
