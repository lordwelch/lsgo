package lslib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/go-kit/kit/log"
)

type LSBHeader struct {
	Signature  [4]byte
	Size       uint32
	Endianness uint32
	Unknown    uint32
	Version    LSMetadata
}

func (lsbh *LSBHeader) Read(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		n   int
		err error
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "header")
	pos, err = r.Seek(0, io.SeekCurrent)

	n, err = r.Read(lsbh.Signature[:])
	if err != nil {
		return err
	}
	l.Log("member", "Signature", "read", n, "start position", pos, "value", fmt.Sprintf("%#x", lsbh.Signature[:]))
	pos += int64(n)
	err = binary.Read(r, binary.LittleEndian, &lsbh.Size)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Size", "read", n, "start position", pos, "value", lsbh.Size)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Endianness)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Endianness", "read", n, "start position", pos, "value", lsbh.Endianness)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Unknown)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Unknown", "read", n, "start position", pos, "value", lsbh.Unknown)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Version.Timestamp)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Timestamp", "read", n, "start position", pos, "value", lsbh.Version.Timestamp)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Version.Major)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Major", "read", n, "start position", pos, "value", lsbh.Version.Major)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Version.Minor)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Minor", "read", n, "start position", pos, "value", lsbh.Version.Minor)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Version.Revision)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Revision", "read", n, "start position", pos, "value", lsbh.Version.Revision)

	err = binary.Read(r, binary.LittleEndian, &lsbh.Version.Build)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version.Build", "read", n, "start position", pos, "value", lsbh.Version.Build)
	pos += int64(n)

	return nil
}

type IdentifierDictionary map[int]string

func ReadLSB(r io.ReadSeeker) (Resource, error) {
	var (
		hdr = &LSBHeader{}
		h   = [4]byte{0x00, 0x00, 0x00, 0x40}
		err error
		d   IdentifierDictionary
		res Resource

		l   log.Logger
		pos int64
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "file")
	pos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "header", "start position", pos)

	err = hdr.Read(r)
	if err != nil {
		return Resource{}, err
	}
	if !(hdr.Signature == [4]byte{'L', 'S', 'F', 'M'} || hdr.Signature == h) {
		return Resource{}, HeaderError{
			Expected: []byte("LSFM"),
			Got:      hdr.Signature[:],
		}
	}

	pos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "string dictionary", "start position", pos)
	d, err = ReadLSBDictionary(r, binary.LittleEndian)
	if err != nil {
		return Resource{}, err
	}

	pos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "Regions", "start position", pos)

	res, err = ReadLSBRegions(r, d, binary.LittleEndian, FileVersion(hdr.Version.Major))
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "dictionary")
	pos, err = r.Seek(0, io.SeekCurrent)

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

		str, err = ReadCString(r, int(stringLength))
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

func ReadLSBRegions(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version FileVersion) (Resource, error) {
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "region")
	pos, err = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, endianness, &regionCount)
	n = 4
	if err != nil {
		return Resource{}, err
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
			return Resource{}, err
		}
		l.Log("member", "key", "read", n, "start position", pos, "value", d[int(key)], "key", key)
		pos += int64(n)
		if regions[i].name, ok = d[int(key)]; !ok {
			return Resource{}, ErrInvalidNameKey
		}
		err = binary.Read(r, endianness, &regions[i].offset)
		n = 4
		if err != nil {
			return Resource{}, err
		}
		l.Log("member", "offset", "read", n, "start position", pos, "value", regions[i].offset)
		pos += int64(n)
	}
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].offset < regions[j].offset
	})
	res := Resource{
		Regions: make([]*Node, 0, regionCount),
	}
	for _, re := range regions {
		var node *Node
		node, err = readLSBNode(r, d, endianness, version, re.offset)
		if err != nil {
			return res, err
		}
		node.RegionName = re.name
		res.Regions = append(res.Regions, node)
	}
	return res, nil
}

func readLSBNode(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version FileVersion, offset uint32) (*Node, error) {
	var (
		key        uint32
		attrCount  uint32
		childCount uint32
		node       = new(Node)
		err        error
		ok         bool

		l   log.Logger
		pos int64
		n   int
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "node")
	pos, err = r.Seek(0, io.SeekCurrent)

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

	node.Attributes = make([]NodeAttribute, int(attrCount))

	for i := range node.Attributes {
		node.Attributes[i], err = readLSBAttribute(r, d, endianness, version)
		if err != nil {
			return node, err
		}
	}

	node.Children = make([]*Node, int(childCount))
	for i := range node.Children {
		node.Children[i], err = readLSBNode(r, d, endianness, version, 0)
		if err != nil {
			return node, err
		}
	}
	return node, nil
}

func readLSBAttribute(r io.ReadSeeker, d IdentifierDictionary, endianness binary.ByteOrder, version FileVersion) (NodeAttribute, error) {
	var (
		key      uint32
		name     string
		attrType uint32
		attr     NodeAttribute
		err      error
		ok       bool
	)
	err = binary.Read(r, endianness, &key)
	if err != nil {
		return attr, err
	}
	if name, ok = d[int(key)]; !ok {
		return attr, ErrInvalidNameKey
	}
	err = binary.Read(r, endianness, &attrType)
	if err != nil {
		return attr, err
	}
	return ReadLSBAttr(r, name, DataType(attrType), endianness, version)
}

func ReadLSBAttr(r io.ReadSeeker, name string, dt DataType, endianness binary.ByteOrder, version FileVersion) (NodeAttribute, error) {
	// LSF and LSB serialize the buffer types differently, so specialized
	// code is added to the LSB and LSf serializers, and the common code is
	// available in BinUtils.ReadAttribute()
	var (
		attr = NodeAttribute{
			Type: dt,
			Name: name,
		}
		err    error
		length uint32

		l   log.Logger
		pos int64
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsb", "part", "attribute")
	pos, err = r.Seek(0, io.SeekCurrent)

	switch dt {
	case DTString, DTPath, DTFixedString, DTLSString: //, DTLSWString:
		var v string
		err = binary.Read(r, endianness, &length)
		if err != nil {
			return attr, err
		}
		v, err = ReadCString(r, int(length))
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTWString:
		panic("Not implemented")

	case DTTranslatedString:
		var v TranslatedString
		v, err = ReadTranslatedString(r, version, 0)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTTranslatedFSString:
		panic("Not implemented")
		var v TranslatedFSString
		// v, err = ReadTranslatedFSString(r, Version)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	case DTScratchBuffer:
		panic("Not implemented")

		v := make([]byte, length)
		_, err = r.Read(v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)

		return attr, err

	default:
		return ReadAttribute(r, name, dt, uint(length), l)
	}
}
