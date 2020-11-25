package lslib

import "errors"

type FileVersion uint32

const (
	/// <summary>
	/// Initial version of the LSF format
	/// </summary>
	VerInitial FileVersion = iota + 1

	/// <summary>
	/// LSF version that added chunked compression for substreams
	/// </summary>
	VerChunkedCompress

	/// <summary>
	/// LSF version that extended the node descriptors
	/// </summary>
	VerExtendedNodes

	/// <summary>
	/// BG3 version, no changes found so far apart from version numbering
	/// </summary>
	VerBG3

	/// <summary>
	/// Latest version supported by this library
	/// </summary>
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
