package lslib

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/pierrec/lz4"
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
		panic(errors.New("Invalid compression flags"))
	}
	return 0
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

func Decompress(compressed io.Reader, compressionFlags byte, chunked bool) io.ReadSeeker {
	switch CompressionMethod(compressionFlags & 0x0f) {
	case CMNone:
		if v, ok := compressed.(io.ReadSeeker); ok {
			return v
		}
		panic(errors.New("compressed must be an io.ReadSeeker if there is no compression"))

	case CMZlib:
		zr, _ := zlib.NewReader(compressed)
		v, _ := ioutil.ReadAll(zr)
		return bytes.NewReader(v)

	case CMLZ4:
		if chunked {
			zr := lz4.NewReader(compressed)
			v, _ := ioutil.ReadAll(zr)
			return bytes.NewReader(v)
		} else {
			panic(errors.New("not implemented"))
			// src, _ := ioutil.ReadAll(compressed)
			// dst := make([]byte, decompressedSize)
			// lz4.UncompressBlock(src, dst)

			// return bytes.NewReader(dst)
		}

	default:
		panic(fmt.Errorf("No decompressor found for this format: %v", compressionFlags))
	}
}

func ReadCString(r io.Reader, length int) (string, error) {
	var err error
	buf := make([]byte, length)
	_, err = r.Read(buf)
	if err != nil {
		return string(buf[:clen(buf)]), err
	}

	if buf[len(buf)-1] != 0 {
		return string(buf[:clen(buf)]), errors.New("string is not null-terminated")
	}

	return string(buf[:clen(buf)]), nil
}

func ReadAttribute(r io.Reader, name string, DT DataType) (NodeAttribute, error) {
	var (
		attr = NodeAttribute{
			Type: DT,
			Name: name,
		}
		err error
	)
	switch DT {
	case DT_None:
		return attr, nil

	case DT_Byte:
		p := make([]byte, 1)
		_, err = r.Read(p)
		attr.Value = p[0]
		return attr, err

	case DT_Short:
		var v int16
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_UShort:
		var v uint16
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_Int:
		var v int32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_UInt:
		var v uint32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_Float:
		var v float32
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_Double:
		var v float64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_IVec2, DT_IVec3, DT_IVec4:
		var col int
		col, err = attr.GetColumns()
		if err != nil {
			return attr, err
		}
		vec := make(Ivec, col)
		for i, _ := range vec {
			var v int32
			err = binary.Read(r, binary.LittleEndian, &v)
			if err != nil {
				return attr, err
			}
			vec[i] = int(v)
		}
		attr.Value = vec
		return attr, nil

	case DT_Vec2, DT_Vec3, DT_Vec4:
		var col int
		col, err = attr.GetColumns()
		if err != nil {
			return attr, err
		}
		vec := make(Vec, col)
		for i, _ := range vec {
			var v float32
			err = binary.Read(r, binary.LittleEndian, &v)
			if err != nil {
				return attr, err
			}
			vec[i] = v
		}
		attr.Value = vec
		return attr, nil

	case DT_Mat2, DT_Mat3, DT_Mat3x4, DT_Mat4x3, DT_Mat4:
		// int columns = attr.GetColumns();
		// int rows = attr.GetRows();
		// var mat = new Matrix(rows, columns);
		// attr.Value = mat;

		// for (int col = 0; col < columns; col++)                        {
		//     for (int row = 0; row < rows; row++)                            {
		//         mat[row, col] = reader.ReadSingle();
		//     }
		// }
		return attr, errors.New("not implemented")

	case DT_Bool:
		var v bool
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_ULongLong:
		var v uint64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_Long, DT_Int64:
		var v int64
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_Int8:
		var v int8
		err = binary.Read(r, binary.LittleEndian, &v)
		attr.Value = v
		return attr, err

	case DT_UUID:
		var v uuid.UUID
		p := make([]byte, 16)
		r.Read(p)
		reverse(p[:4])
		reverse(p[4:6])
		reverse(p[6:8])
		v, err = uuid.FromBytes(p)
		attr.Value = v
		return attr, err

	default:
		// Strings are serialized differently for each file format and should be
		// handled by the format-specific ReadAttribute()
		return attr, fmt.Errorf("ReadAttribute() not implemented for type %v", DT)
	}

	return attr, nil
}

// LimitReader returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func LimitReadSeeker(r io.ReadSeeker, n int64) io.ReadSeeker { return &LimitedReadSeeker{r, n} }

// A LimitedReader reads from R but limits the amount of
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
