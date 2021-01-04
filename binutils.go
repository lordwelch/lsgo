package lsgo

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/pierrec/lz4/v4"
	"gonum.org/v1/gonum/mat"
)

func reverse(nums []byte) {
	i := 0
	j := len(nums) - 1
	for i < j {
		nums[i], nums[j] = nums[j], nums[i]
		i++
		j--
	}
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

func CompressionFlagsToMethod(flags byte) CompressionMethod {
	switch CompressionMethod(flags & 0x0f) {
	case CMNone:
		return CMNone

	case CMZlib:
		return CMZlib

	case CMLZ4:
		return CMLZ4

	default:
		return CMInvalid
	}
}

func CompressionFlagsToLevel(flags byte) CompressionLevel {
	switch CompressionLevel(flags & 0xf0) {
	case FastCompression:
		return FastCompression

	case DefaultCompression:
		return DefaultCompression

	case MaxCompression:
		return MaxCompression

	default:
		panic(errors.New("invalid compression flags"))
	}
	// return 0
}

func MakeCompressionFlags(method CompressionMethod, level CompressionLevel) int {
	if method == CMNone || method == CMInvalid {
		return 0
	}

	var flags int = 0
	if method == CMZlib {
		flags = 0x1
	} else if method == CMLZ4 {
		flags = 0x2
	}

	return flags | int(level)
}

func Decompress(compressed io.Reader, uncompressedSize int, compressionFlags byte, chunked bool) io.ReadSeeker {
	switch CompressionMethod(compressionFlags & 0x0f) {
	case CMNone:
		// logger.Println("No compression")
		if v, ok := compressed.(io.ReadSeeker); ok {
			return v
		}
		panic(errors.New("compressed must be an io.ReadSeeker if there is no compression"))

	case CMZlib:
		// logger.Println("zlib compression")
		zr, _ := zlib.NewReader(compressed)
		v, _ := ioutil.ReadAll(zr)
		return bytes.NewReader(v)

	case CMLZ4:
		if chunked {
			// logger.Println("lz4 stream compressed")
			zr := lz4.NewReader(compressed)
			p := make([]byte, uncompressedSize)
			_, err := zr.Read(p)
			if err != nil {
				panic(err)
			}
			return bytes.NewReader(p)
		}
		// logger.Println("lz4 block compressed")
		// panic(errors.New("not implemented"))
		src, _ := ioutil.ReadAll(compressed)
		// logger.Println(len(src))
		dst := make([]byte, uncompressedSize*2)
		_, err := lz4.UncompressBlock(src, dst)
		if err != nil {
			panic(err)
		}

		return bytes.NewReader(dst)

	default:
		panic(fmt.Errorf("no decompressor found for this format: %v", compressionFlags))
	}
}

func ReadCString(r io.Reader, length int) (string, error) {
	var err error
	buf := make([]byte, length)
	_, err = r.Read(buf)
	if err != nil {
		return string(buf[:clen(buf)]), err
	}

	return string(buf[:clen(buf)]), nil
}

func ReadAttribute(r io.ReadSeeker, name string, DT DataType, length uint, l log.Logger) (NodeAttribute, error) {
	var (
		attr = NodeAttribute{
			Type: DT,
			Name: name,
		}
		err error

		pos int64
		n   int
	)
	pos, err = r.Seek(0, io.SeekCurrent)

	switch DT {
	case DTNone:

		l.Log("member", name, "read", length, "start position", pos, "value", nil)

		return attr, nil

	case DTByte:
		p := make([]byte, 1)
		n, err = r.Read(p)
		attr.Value = p[0]

		l.Log("member", name, "read", n, "start position", pos, "value", attr.Value)

		return attr, err

	case DTShort:
		var v int16
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTUShort:
		var v uint16
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTInt:
		var v int32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTUInt:
		var v uint32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTFloat:
		var v float32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTDouble:
		var v float64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTIVec2, DTIVec3, DTIVec4:
		var col int
		col, err = attr.GetColumns()
		if err != nil {
			return attr, err
		}
		vec := make(Ivec, col)
		for i := range vec {
			var v int32
			err = binary.Read(r, binary.LittleEndian, &v)
			if err != nil {
				return attr, err
			}
			vec[i] = int(v)
		}
		attr.Value = vec

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, nil

	case DTVec2, DTVec3, DTVec4:
		var col int
		col, err = attr.GetColumns()
		if err != nil {
			return attr, err
		}
		vec := make(Vec, col)
		for i := range vec {
			var v float32
			err = binary.Read(r, binary.LittleEndian, &v)
			if err != nil {
				return attr, err
			}
			vec[i] = float64(v)
		}
		attr.Value = vec

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, nil

	case DTMat2, DTMat3, DTMat3x4, DTMat4x3, DTMat4:
		var (
			row int
			col int
		)
		col, err = attr.GetColumns()
		if err != nil {
			return attr, err
		}
		row, err = attr.GetRows()
		if err != nil {
			return attr, err
		}
		vec := make(Vec, col*row)

		for c := 0; c < col; c++ {
			for ro := 0; ro < row; ro++ {
				var v float32
				err = binary.Read(r, binary.LittleEndian, &v)
				if err != nil {
					return attr, err
				}
				vec[ro*col+c] = float64(v)
			}
		}
		attr.Value = (*Mat)(mat.NewDense(row, col, []float64(vec)))

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, nil

	case DTBool:
		var v bool
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTULongLong:
		var v uint64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTLong, DTInt64:
		var v int64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTInt8:
		var v int8
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTUUID:
		var v uuid.UUID
		p := make([]byte, 16)
		n, err = r.Read(p)
		if err != nil {
			return attr, err
		}
		reverse(p[:4])
		reverse(p[4:6])
		reverse(p[6:8])
		v, err = uuid.FromBytes(p)
		attr.Value = v

		l.Log("member", name, "read", n, "start position", pos, "value", attr.Value)

		return attr, err

	default:
		// Strings are serialized differently for each file format and should be
		// handled by the format-specific ReadAttribute()
		// pretty.Log(attr)
		return attr, fmt.Errorf("readAttribute() not implemented for type %v", DT)
	}

	// return attr, nil
}

// LimitReadSeeker returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func LimitReadSeeker(r io.ReadSeeker, n int64) io.ReadSeeker { return &LimitedReadSeeker{r, n} }

// A LimitedReadSeeker reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReadSeeker struct {
	R io.ReadSeeker // underlying reader
	N int64         // max bytes remaining
}

func (l *LimitedReadSeeker) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

func (l *LimitedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		n, err := l.R.Seek(0, io.SeekCurrent)
		if err != nil {
			return n, err
		}
		l.N += n - offset
		return l.R.Seek(offset, whence)

	case io.SeekEnd:
		n, err := l.R.Seek(l.N, io.SeekCurrent)
		if err != nil {
			return n, err
		}
		l.N = 0 - offset
		return l.R.Seek(offset, io.SeekCurrent)

	case io.SeekCurrent:
		l.N -= offset
		return l.R.Seek(offset, whence)
	default:
		return -1, io.ErrNoProgress
	}
}
