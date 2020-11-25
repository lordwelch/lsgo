package lslib

import (
	"encoding/binary"
	"io"

	"github.com/go-kit/kit/log"
)

type LSBHeader struct {
	Signature  [4]byte
	Size       uint32
	Endianness uint32
	Unknown    uint32
	Version    struct {
		Major    uint32
		Minor    uint32
		Build    uint32
		Revision uint32
	}
}

type LSBRegion struct {
	name string
	offset
}

type IdentifierDictionary map[int]string

func ReadLSBDictionary(r io.Reader, endianness binary.ByteOrder) (IdentifierDictionary, error) {
	var (
		dict IdentifierDictionary
		size uint32
		err  error
	)
	err = binary.Read(r, endianness, &size)
	if err != nil {
		return nil, err
	}
	dict = make(IdentifierDictionary, size)
	for i := 0; i < int(size); i++ {
		var (
			stringlength uint32
			key          uint32
			str          string
		)
		err = binary.Read(r, endianness, &stringlength)
		if err != nil {
			return dict, err
		}
		str, err = ReadCString(r, int(stringlength))
		if err != nil {
			return dict, err
		}
		err = binary.Read(r, endianness, &key)
		if err != nil {
			return dict, err
		}
		dict[key] = str
	}
}

func ReadLSBRegions(r io.Reader, d IdentifierDictionary, endianness binary.ByteOrder) (Resource, error) {
	var (
		nodes []struct {
			node   *Node
			offset uint32
		}
		nodeCount uint32
		err       error
	)

	err = binary.Read(r, endianness, &nodeCount)
	if err != nil {
		return dict, err
	}
	nodes = make([]struct{ Node, offset uint32 }, nodeCount)
	for _, n := range nodes {
		var (
			key uint32
			ok  bool
		)
		err = binary.Read(r, endianness, &key)
		if err != nil {
			return dict, err
		}
		n.node = new(Node)
		if n.node.Name, ok = d[int(key)]; !ok {
			return Resource{}, ErrInvalidNameKey
		}
		err = binary.Read(r, endianness, &n.offset)
		if err != nil {
			return dict, err
		}
	}
	// TODO: Sort
	for _, n := range nodes {
		var (
			key        uint32
			attrCount  uint32
			childCount uint32
		)
		// TODO: Check offset

		err = binary.Read(r, endianness, &key)
		if err != nil {
			return dict, err
		}
		// if keyV, ok := d[int(key)]; !ok {
		// 	return Resource{}, ErrKeyDoesNotMatch
		// }

		err = binary.Read(r, endianness, &attrCount)
		if err != nil {
			return dict, err
		}
		n.node.Attributes = make([]NodeAttribute, int(attrCount))
		err = binary.Read(r, endianness, &nodeCount)
		if err != nil {
			return dict, err
		}

	}
}

func readLSBAttribute(r io.Reader) (NodeAttribute, err) {
	var (
		key      uint32
		name     string
		attrType uint32
		attr     NodeAttribute
	)
	err = binary.Read(r, endianness, &key)
	if err != nil {
		return dict, err
	}
	if name, ok = d[int(key)]; !ok {
		return Resource{}, ErrInvalidNameKey
	}
	err = binary.Read(r, endianness, &attrType)
	if err != nil {
		return dict, err
	}
	ReadLSBAttribute(r, name, DataType(attrType))
}

func ReadLSBAttr(r io.ReadSeeker, name string, DT DataType) (NodeAttribute, error) {
	// LSF and LSB serialize the buffer types differently, so specialized
	// code is added to the LSB and LSf serializers, and the common code is
	// available in BinUtils.ReadAttribute()
	var (
		attr = NodeAttribute{
			Type: DT,
			Name: name,
		}
		err error

		l   log.Logger
		pos int64
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "attribute")
	pos, err = r.Seek(0, io.SeekCurrent)

	switch DT {
	case DT_String, DT_Path, DT_FixedString, DT_LSString, DT_WString, DT_LSWString:
		var v string
		v, err = ReadCString(r, int(length))
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case DT_TranslatedString:
		var v TranslatedString
		v, err = ReadTranslatedString(r, Version, EngineVersion)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case DT_TranslatedFSString:
		var v TranslatedFSString
		v, err = ReadTranslatedFSString(r, Version)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case DT_ScratchBuffer:

		v := make([]byte, length)
		_, err = r.Read(v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	default:
		return ReadAttribute(r, name, DT, length, l)
	}
}
