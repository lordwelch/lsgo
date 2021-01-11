package lsb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"

	"git.narnian.us/lordwelch/lsgo"

	"github.com/go-kit/kit/log"
)

const (
	Signature       = "LSFM"
	PreBG3Signature = "\x00\x00\x00\x40"
)

type Header struct {
	Signature  [4]byte
	Size       uint32
	Endianness uint32
	Unknown    uint32
	Version    lsgo.LSMetadata
}

func (h *Header) Read(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		n   int
		err error
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "header")
	pos, _ = r.Seek(0, io.SeekCurrent)

	n, err = r.Read(h.Signature[:])
	if err != nil {
		return err
	}
	l.Log("member", "Signature", "read", n, "start position", pos, "value", fmt.Sprintf("%#x", h.Signature[:]))
	pos += int64(n)
	err = binary.Read(r, binary.LittleEndian, &h.Size)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Size", "read", n, "start position", pos, "value", h.Size)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Endianness)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Endianness", "read", n, "start position", pos, "value", h.Endianness)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Unknown)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Unknown", "read", n, "start position", pos, "value", h.Unknown)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Version.Timestamp)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Timestamp", "read", n, "start position", pos, "value", h.Version.Timestamp)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Version.Major)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Major", "read", n, "start position", pos, "value", h.Version.Major)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Version.Minor)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Minor", "read", n, "start position", pos, "value", h.Version.Minor)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Version.Revision)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Revision", "read", n, "start position", pos, "value", h.Version.Revision)

	err = binary.Read(r, binary.LittleEndian, &h.Version.Build)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Build", "read", n, "start position", pos, "value", h.Version.Build)
	pos += int64(n)

	return nil
}

type IdentifierDictionary map[int]string

func Read(r io.ReadSeeker) (lsgo.Resource, error) {
	var (
		hdr = &Header{}
		err error
		d   IdentifierDictionary
		res lsgo.Resource

		l   log.Logger
		pos int64
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "file")
	pos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "header", "start position", pos)

	err = hdr.Read(r)
	if err != nil {
		return lsgo.Resource{}, err
	}
	if !(string(hdr.Signature[:]) == Signature || string(hdr.Signature[:]) == PreBG3Signature) {
		return lsgo.Resource{}, lsgo.HeaderError{
			Expected: Signature,
			Got:      hdr.Signature[:],
		}
	}

	pos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "string dictionary", "start position", pos)
	d, err = ReadLSBDictionary(r, binary.LittleEndian)
	if err != nil {
		return lsgo.Resource{}, err
	}

	pos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "Regions", "start position", pos)

	res, err = ReadLSBRegions(r, d, binary.LittleEndian, lsgo.FileVersion(hdr.Version.Major))
	res.Metadata = hdr.Version
	return res, err
}

func ReadLSBDictionary(r io.ReadSeeker, endianness binary.ByteOrder) (IdentifierDictionary, error) {
	var (
		dict   IdentifierDictionary
		length uint32
		err    error

		l   log.Logger
		pos int64
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "dictionary")
	pos, _ = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, endianness, &length)
	n = 4
	if err != nil {
		return nil, err
	}
	l.Log("member", "length", "read", n, "start position", pos, "value", length)
	pos += int64(n)

	dict = make(IdentifierDictionary, length)
	for i := 0; i < int(length); i++ {
		var (
			stringLength uint32
			key          uint32
			str          string
		)
		err = binary.Read(r, endianness, &stringLength)
		n = 4
		if err != nil {
			return dict, err
		}
		l.Log("member", "stringLength", "read", n, "start position", pos, "value", stringLength)
		pos += int64(n)

		str, err = lsgo.ReadCString(r, int(stringLength))
		n += int(stringLength)
		if err != nil {
			return dict, err
		}
		l.Log("member", "str", "read", n, "start position", pos, "value", str)
		pos += int64(n)

		err = binary.Read(r, endianness, &key)
		n = 4
		if err != nil {
			return dict, err
		}
		dict[int(key)] = str
		l.Log("member", "key", "read", n, "start position", pos, "value", key)
		pos += int64(n)
	}
	return dict, nil
}

func ReadLSBRegions(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version lsgo.FileVersion) (lsgo.Resource, error) {
	var (
		regions []struct {
			name   string
			offset uint32
		}
		regionCount uint32
		err         error

		l   log.Logger
		pos int64
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "region")
	pos, _ = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, endianness, &regionCount)
	n = 4
	if err != nil {
		return lsgo.Resource{}, err
	}
	l.Log("member", "regionCount", "read", n, "start position", pos, "value", regionCount)
	pos += int64(n)

	regions = make([]struct {
		name   string
		offset uint32
	}, regionCount)
	for i := range regions {
		var (
			key uint32
			ok  bool
		)
		err = binary.Read(r, endianness, &key)
		n = 4
		if err != nil {
			return lsgo.Resource{}, err
		}
		l.Log("member", "key", "read", n, "start position", pos, "value", d[int(key)], "key", key)
		pos += int64(n)
		if regions[i].name, ok = d[int(key)]; !ok {
			return lsgo.Resource{}, lsgo.ErrInvalidNameKey
		}
		err = binary.Read(r, endianness, &regions[i].offset)
		n = 4
		if err != nil {
			return lsgo.Resource{}, err
		}
		l.Log("member", "offset", "read", n, "start position", pos, "value", regions[i].offset)
		pos += int64(n)
	}
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].offset < regions[j].offset
	})
	res := lsgo.Resource{
		Regions: make([]*lsgo.Node, 0, regionCount),
	}
	for _, re := range regions {
		var node *lsgo.Node
		node, err = readLSBNode(r, d, endianness, version, re.offset)
		if err != nil {
			return res, err
		}
		node.RegionName = re.name
		res.Regions = append(res.Regions, node)
	}
	return res, nil
}

func readLSBNode(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version lsgo.FileVersion, offset uint32) (*lsgo.Node, error) {
	var (
		key        uint32
		attrCount  uint32
		childCount uint32
		node       = new(lsgo.Node)
		err        error
		ok         bool

		l   log.Logger
		pos int64
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "node")
	pos, _ = r.Seek(0, io.SeekCurrent)

	if pos != int64(offset) && offset != 0 {
		panic("shit")
	}

	err = binary.Read(r, endianness, &key)
	n = 4
	if err != nil {
		return nil, err
	}
	l.Log("member", "key", "read", n, "start position", pos, "value", d[int(key)], "key", key)
	pos += int64(n)

	if node.Name, ok = d[int(key)]; !ok {
		return nil, errors.New("node id key is invalid")
	}

	err = binary.Read(r, endianness, &attrCount)
	n = 4
	if err != nil {
		return node, err
	}
	l.Log("member", "attrCount", "read", n, "start position", pos, "value", attrCount)
	pos += int64(n)

	err = binary.Read(r, endianness, &childCount)
	n = 4
	if err != nil {
		return node, err
	}
	l.Log("member", "childCount", "read", n, "start position", pos, "value", childCount)

	node.Attributes = make([]lsgo.NodeAttribute, int(attrCount))

	for i := range node.Attributes {
		node.Attributes[i], err = readLSBAttribute(r, d, endianness, version)
		if err != nil {
			return node, err
		}
	}

	node.Children = make([]*lsgo.Node, int(childCount))
	for i := range node.Children {
		node.Children[i], err = readLSBNode(r, d, endianness, version, 0)
		if err != nil {
			return node, err
		}
	}
	return node, nil
}

func readLSBAttribute(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version lsgo.FileVersion) (lsgo.NodeAttribute, error) {
	var (
		key      uint32
		name     string
		attrType uint32
		attr     lsgo.NodeAttribute
		err      error
		ok       bool
	)
	err = binary.Read(r, endianness, &key)
	if err != nil {
		return attr, err
	}
	if name, ok = d[int(key)]; !ok {
		return attr, lsgo.ErrInvalidNameKey
	}
	err = binary.Read(r, endianness, &attrType)
	if err != nil {
		return attr, err
	}
	return ReadLSBAttr(r, name, lsgo.DataType(attrType), endianness, version)
}

func ReadLSBAttr(r io.ReadSeeker, name string, dt lsgo.DataType, endianness binary.ByteOrder, version lsgo.FileVersion) (lsgo.NodeAttribute, error) {
	// LSF and LSB serialize the buffer types differently, so specialized
	// code is added to the LSB and LSf serializers, and the common code is
	// available in BinUtils.ReadAttribute()
	var (
		attr = lsgo.NodeAttribute{
			Type: dt,
			Name: name,
		}
		err    error
		length uint32

		l   log.Logger
		pos int64
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsb", "part", "attribute")
	pos, _ = r.Seek(0, io.SeekCurrent)

	switch dt {
	case lsgo.DTString, lsgo.DTPath, lsgo.DTFixedString, lsgo.DTLSString: // DTLSWString:
		var v string
		err = binary.Read(r, endianness, &length)
		if err != nil {
			return attr, err
		}
		v, err = lsgo.ReadCString(r, int(length))
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case lsgo.DTWString:
		panic("Not implemented")

	case lsgo.DTTranslatedString:
		var v lsgo.TranslatedString
		v, err = lsgo.ReadTranslatedString(r, version, 0)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case lsgo.DTTranslatedFSString:
		panic("Not implemented")
		var v lsgo.TranslatedFSString
		// v, err = ReadTranslatedFSString(r, Version)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case lsgo.DTScratchBuffer:
		panic("Not implemented")

		v := make([]byte, length)
		_, err = r.Read(v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	default:
		return lsgo.ReadAttribute(r, name, dt, uint(length), l)
	}
}

func init() {
	lsgo.RegisterFormat("lsb", Signature, Read)
	lsgo.RegisterFormat("lsb", PreBG3Signature, Read)
}
