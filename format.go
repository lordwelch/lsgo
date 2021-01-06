// Adapted from the image package
package lsgo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// ErrFormat indicates that decoding encountered an unknown format.
var ErrFormat = errors.New("lsgo: unknown format")

// A format holds an image format's name, magic header and how to decode it.
type format struct {
	name, magic string
	decode      func(io.ReadSeeker) (Resource, error)
}

// Formats is the list of registered formats.
var (
	formatsMu     sync.Mutex
	atomicFormats atomic.Value
)

// RegisterFormat registers an image format for use by Decode.
// Name is the name of the format, like "jpeg" or "png".
// Magic is the magic prefix that identifies the format's encoding. The magic
// string can contain "?" wildcards that each match any one byte.
// Decode is the function that decodes the encoded image.
// DecodeConfig is the function that decodes just its configuration.
func RegisterFormat(name, magic string, decode func(io.ReadSeeker) (Resource, error)) {
	formatsMu.Lock()
	formats, _ := atomicFormats.Load().([]format)
	atomicFormats.Store(append(formats, format{name, magic, decode}))
	formatsMu.Unlock()
}

// Match reports whether magic matches b. Magic may contain "?" wildcards.
func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// Sniff determines the format of r's data.
func sniff(r io.ReadSeeker) format {
	var (
		b   []byte = make([]byte, 4)
		err error
	)
	formats, _ := atomicFormats.Load().([]format)
	for _, f := range formats {
		if len(b) < len(f.magic) {
			fmt.Fprintln(os.Stderr, f.magic)
			b = make([]byte, len(f.magic))
		}
		_, err = r.Read(b)
		r.Seek(0, io.SeekStart)
		if err == nil && match(f.magic, b) {
			return f
		}
	}
	return format{}
}

func Decode(r io.ReadSeeker) (Resource, string, error) {
	f := sniff(r)
	if f.decode == nil {
		return Resource{}, "", ErrFormat
	}
	m, err := f.decode(r)
	return m, f.name, err
}

func SupportedFormat(signature []byte) bool {
	formats, _ := atomicFormats.Load().([]format)
	for _, f := range formats {
		if match(f.magic, signature) {
			return true
		}
	}
	return false
}
