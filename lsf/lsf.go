package lsf

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"git.narnian.us/lordwelch/lsgo"

	"github.com/go-kit/kit/log"
)

const Signature = "LSOF"

type Header struct {
	// LSOF file signature
	Signature [4]byte

	// Version of the LSOF file D:OS EE is version 1/2, D:OS 2 is version 3
	Version lsgo.FileVersion

	// Possibly version number? (major, minor, rev, build)
	EngineVersion uint32

	// Total uncompressed size of the string hash table
	StringsUncompressedSize uint32

	// Compressed size of the string hash table
	StringsSizeOnDisk uint32

	// Total uncompressed size of the node list
	NodesUncompressedSize uint32

	// Compressed size of the node list
	NodesSizeOnDisk uint32

	// Total uncompressed size of the attribute list
	AttributesUncompressedSize uint32

	// Compressed size of the attribute list
	AttributesSizeOnDisk uint32

	// Total uncompressed size of the raw value buffer
	ValuesUncompressedSize uint32

	// Compressed size of the raw value buffer
	ValuesSizeOnDisk uint32

	// Uses the same format as packages (see BinUtils.MakeCompressionFlags)
	CompressionFlags byte

	// Possibly unused, always 0
	Unknown2 byte
	Unknown3 uint16

	// Extended node/attribute format indicator, 0 for V2, 0/1 for V3
	Extended uint32
}

func (h *Header) Read(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		n   int
		err error
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "header")
	pos, _ = r.Seek(0, io.SeekCurrent)
	n, err = r.Read(h.Signature[:])
	if err != nil {
		return err
	}
	l.Log("member", "Signature", "read", n, "start position", pos, "value", string(h.Signature[:]))
	pos += int64(n)
	err = binary.Read(r, binary.LittleEndian, &h.Version)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version", "read", n, "start position", pos, "value", h.Version)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.EngineVersion)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "EngineVersion", "read", n, "start position", pos, "value", fmt.Sprintf("%d.%d.%d.%d", (h.EngineVersion&0xf0000000)>>28, (h.EngineVersion&0xf000000)>>24, (h.EngineVersion&0xff0000)>>16, (h.EngineVersion&0xffff)))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.StringsUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "StringsUncompressedSize", "read", n, "start position", pos, "value", h.StringsUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.StringsSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "StringsSizeOnDisk", "read", n, "start position", pos, "value", h.StringsSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.NodesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NodesUncompressedSize", "read", n, "start position", pos, "value", h.NodesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.NodesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NodesSizeOnDisk", "read", n, "start position", pos, "value", h.NodesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.AttributesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "AttributesUncompressedSize", "read", n, "start position", pos, "value", h.AttributesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.AttributesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "AttributesSizeOnDisk", "read", n, "start position", pos, "value", h.AttributesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.ValuesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "ValuesUncompressedSize", "read", n, "start position", pos, "value", h.ValuesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.ValuesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "ValuesSizeOnDisk", "read", n, "start position", pos, "value", h.ValuesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.CompressionFlags)
	n = 1
	if err != nil {
		return err
	}
	l.Log("member", "CompressionFlags", "read", n, "start position", pos, "value", h.CompressionFlags)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Unknown2)
	n = 1
	if err != nil {
		return err
	}
	l.Log("member", "Unknown2", "read", n, "start position", pos, "value", h.Unknown2)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Unknown3)
	n = 2
	if err != nil {
		return err
	}
	l.Log("member", "Unknown3", "read", n, "start position", pos, "value", h.Unknown3)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &h.Extended)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Extended", "read", n, "start position", pos, "value", h.Extended)
	pos += int64(n)

	if !h.IsCompressed() {
		h.NodesSizeOnDisk = h.NodesUncompressedSize
		h.AttributesSizeOnDisk = h.AttributesUncompressedSize
		h.StringsSizeOnDisk = h.StringsUncompressedSize
		h.ValuesSizeOnDisk = h.ValuesUncompressedSize
	}
	return nil
}

func (h Header) IsCompressed() bool {
	return lsgo.CompressionFlagsToMethod(h.CompressionFlags) != lsgo.CMNone && lsgo.CompressionFlagsToMethod(h.CompressionFlags) != lsgo.CMInvalid
}

type NodeEntry struct {
	Long bool

	// (16-bit MSB: index into name hash table, 16-bit LSB: offset in hash chain)
	NameHashTableIndex uint32

	// (-1: node has no attributes)
	FirstAttributeIndex int32

	// (-1: this node is a root region)
	ParentIndex int32

	// (-1: this is the last node)
	NextSiblingIndex int32
}

func (ne *NodeEntry) Read(r io.ReadSeeker) error {
	if ne.Long {
		return ne.readLong(r)
	}
	return ne.readShort(r)
}

func (ne *NodeEntry) readShort(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		err error
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "short node")
	pos, _ = r.Seek(0, io.SeekCurrent)
	err = binary.Read(r, binary.LittleEndian, &ne.NameHashTableIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", strconv.Itoa(ne.NameIndex())+" "+strconv.Itoa(ne.NameOffset()))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.FirstAttributeIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", ne.FirstAttributeIndex)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.ParentIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", ne.ParentIndex)
	pos += int64(n)
	return nil
}

func (ne *NodeEntry) readLong(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		err error
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "long node")
	pos, _ = r.Seek(0, io.SeekCurrent)
	err = binary.Read(r, binary.LittleEndian, &ne.NameHashTableIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", strconv.Itoa(ne.NameIndex())+" "+strconv.Itoa(ne.NameOffset()))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.ParentIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "ParentIndex", "read", n, "start position", pos, "value", ne.ParentIndex)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.NextSiblingIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NextSiblingIndex", "read", n, "start position", pos, "value", ne.NextSiblingIndex)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.FirstAttributeIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "FirstAttributeIndex", "read", n, "start position", pos, "value", ne.FirstAttributeIndex)
	pos += int64(n)

	return nil
}

func (ne NodeEntry) NameIndex() int {
	return int(ne.NameHashTableIndex >> 16)
}

func (ne NodeEntry) NameOffset() int {
	return int(ne.NameHashTableIndex & 0xffff)
}

// Processed node information for a node in the LSF file
type NodeInfo struct {

	// (-1: this node is a root region)
	ParentIndex int

	// Index into name hash table
	NameIndex int

	// Offset in hash chain
	NameOffset int

	// (-1: node has no attributes)
	FirstAttributeIndex int
}

// attribute extension in the LSF file
type AttributeEntry struct {
	Long bool

	// (16-bit MSB: index into name hash table, 16-bit LSB: offset in hash chain)
	NameHashTableIndex uint32

	// 26-bit MSB: Length of this attribute
	TypeAndLength uint32

	// Note: These indexes are assigned seemingly arbitrarily, and are not necessarily indices into the node list
	NodeIndex int32

	// Note: These indexes are assigned seemingly arbitrarily, and are not necessarily indices into the node list
	NextAttributeIndex int32

	// Absolute position of attribute value in the value stream
	Offset uint32
}

func (ae *AttributeEntry) Read(r io.ReadSeeker) error {
	if ae.Long {
		return ae.readLong(r)
	}
	return ae.readShort(r)
}

func (ae *AttributeEntry) readShort(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		err error
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "short attribute")
	pos, _ = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, binary.LittleEndian, &ae.NameHashTableIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", strconv.Itoa(ae.NameIndex())+" "+strconv.Itoa(ae.NameOffset()))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ae.TypeAndLength)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "TypeAndLength", "read", n, "start position", pos, "value", ae.TypeAndLength)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ae.NodeIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NodeIndex", "read", n, "start position", pos, "value", ae.NodeIndex)
	pos += int64(n)

	return nil
}

func (ae *AttributeEntry) readLong(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		err error
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "long attribute")
	pos, _ = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, binary.LittleEndian, &ae.NameHashTableIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", strconv.Itoa(ae.NameIndex())+" "+strconv.Itoa(ae.NameOffset()))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ae.TypeAndLength)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "TypeAndLength", "read", n, "start position", pos, "value", ae.TypeAndLength)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ae.NextAttributeIndex)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NextAttributeIndex", "read", n, "start position", pos, "value", ae.NextAttributeIndex)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ae.Offset)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Offset", "read", n, "start position", pos, "value", ae.Offset)
	pos += int64(n)

	return nil
}

// Index into name hash table
func (ae AttributeEntry) NameIndex() int {
	return int(ae.NameHashTableIndex >> 16)
}

// Offset in hash chain
func (ae AttributeEntry) NameOffset() int {
	return int(ae.NameHashTableIndex & 0xffff)
}

// Type of this attribute (see NodeAttribute.DataType)
func (ae AttributeEntry) TypeID() lsgo.DataType {
	return lsgo.DataType(ae.TypeAndLength & 0x3f)
}

// Length of this attribute
func (ae AttributeEntry) Len() int {
	return int(ae.TypeAndLength >> 6)
}

type AttributeInfo struct {
	V2 bool

	// Index into name hash table
	NameIndex int

	// Offset in hash chain
	NameOffset int

	// Type of this attribute (see NodeAttribute.DataType)
	TypeID lsgo.DataType

	// Length of this attribute
	Length uint

	// Absolute position of attribute data in the values section
	DataOffset uint

	// (-1: this is the last attribute)
	NextAttributeIndex int
}

// extract to lsf package
func ReadNames(r io.ReadSeeker) ([][]string, error) {
	var (
		numHashEntries uint32
		err            error
		names          [][]string

		l   log.Logger
		pos int64
		n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "names")
	pos, _ = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, binary.LittleEndian, &numHashEntries)
	n = 4
	if err != nil {
		return nil, err
	}
	l.Log("member", "numHashEntries", "read", n, "start position", pos, "value", numHashEntries)
	pos += int64(n)

	names = make([][]string, int(numHashEntries))
	for i := range names {
		var numStrings uint16

		err = binary.Read(r, binary.LittleEndian, &numStrings)
		n = 4
		if err != nil {
			return nil, err
		}
		l.Log("member", "numStrings", "read", n, "start position", pos, "value", numStrings)
		pos += int64(n)

		hash := make([]string, int(numStrings))
		for x := range hash {
			var (
				nameLen uint16
				name    []byte
			)
			err = binary.Read(r, binary.LittleEndian, &nameLen)
			n = 2
			if err != nil {
				return nil, err
			}
			l.Log("member", "nameLen", "read", n, "start position", pos, "value", nameLen)
			pos += int64(n)

			name = make([]byte, nameLen)

			n, err = r.Read(name)
			if err != nil {
				return nil, err
			}
			l.Log("member", "name", "read", n, "start position", pos, "value", name)
			pos += int64(n)

			hash[x] = string(name)
		}
		names[i] = hash
	}
	return names, nil
}

func readNodeInfo(r io.ReadSeeker, longNodes bool) ([]NodeInfo, error) {
	var (
		nodes []NodeInfo
		err   error
	)
	index := 0

	for err == nil {
		var node NodeInfo

		item := &NodeEntry{Long: longNodes}
		err = item.Read(r)

		node.FirstAttributeIndex = int(item.FirstAttributeIndex)
		node.NameIndex = item.NameIndex()
		node.NameOffset = item.NameOffset()
		node.ParentIndex = int(item.ParentIndex)

		nodes = append(nodes, node)
		index++
	}
	return nodes[:len(nodes)-1], err
}

// Reads the attribute headers for the LSOF resource
// <param name="s">Stream to read the attribute headers from</param>
func readAttributeInfo(r io.ReadSeeker, long bool) []AttributeInfo {
	// var rawAttributes = new List<AttributeEntryV2>();

	var (
		prevAttributeRefs      = make(map[int]int)
		dataOffset        uint = 0
		index                  = 0
		nextAttrIndex     int  = -1
		attributes        []AttributeInfo
		err               error
	)
	for err == nil {
		attribute := &AttributeEntry{Long: long}
		err = attribute.Read(r)
		if err != nil {
			break
		}

		if long {
			dataOffset = uint(attribute.Offset)
			nextAttrIndex = int(attribute.NextAttributeIndex)
		}

		resolved := AttributeInfo{
			NameIndex:          attribute.NameIndex(),
			NameOffset:         attribute.NameOffset(),
			TypeID:             attribute.TypeID(),
			Length:             uint(attribute.Len()),
			DataOffset:         dataOffset,
			NextAttributeIndex: nextAttrIndex,
		}

		if !long {
			// get index of previous attribute for node
			if indexOfLastAttr, ok := prevAttributeRefs[int(attribute.NodeIndex)]; ok { // previous attribute exists for current node set the next attribute index for the previous node to this attribute
				attributes[indexOfLastAttr].NextAttributeIndex = index
			}
			// set the previous attribute of this node to the current attribute, we are done with it and at the end of the loop
			dataOffset += resolved.Length

			prevAttributeRefs[int(attribute.NodeIndex)] = index
		}

		attributes = append(attributes, resolved)
		index++
	}
	return attributes
	// }

	// Console.WriteLine(" ----- DUMP OF ATTRIBUTE REFERENCES -----");
	// for (int i = 0; i < prevAttributeRefs.Count; i++)
	// {
	//     Console.WriteLine(String.Format("Node {0}: last attribute {1}", i, prevAttributeRefs[i]));
	// }

	// Console.WriteLine(" ----- DUMP OF V2 ATTRIBUTE TABLE -----");
	// for (int i = 0; i < lsfr.Attributes.Count; i++)
	// {
	//     var resolved = lsfr.Attributes[i];
	//     var attribute = rawAttributes[i];

	//     var debug = String.Format(
	//         "{0}: {1} (offset {2:X}, typeId {3}, nextAttribute {4}, node {5})",
	//         i, Names[resolved.NameIndex][resolved.NameOffset], resolved.DataOffset,
	//         resolved.TypeId, resolved.NextAttributeIndex, attribute.NodeIndex
	//     );
	//     Console.WriteLine(debug);
	// }
}

func Read(r io.ReadSeeker) (lsgo.Resource, error) {
	var (
		err error

		// Static string hash map
		names [][]string

		// Preprocessed list of nodes (structures)
		nodeInfo []NodeInfo

		// Preprocessed list of node attributes
		attributeInfo []AttributeInfo

		// Node instances
		nodeInstances []*lsgo.Node
	)
	var (
		l         log.Logger
		pos, npos int64
		// n   int
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "file")
	pos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "header", "start position", pos)

	hdr := &Header{}
	err = hdr.Read(r)
	if err != nil || (string(hdr.Signature[:]) != Signature) {
		return lsgo.Resource{}, lsgo.HeaderError{Expected: Signature, Got: hdr.Signature[:]}
	}

	if hdr.Version < lsgo.VerInitial || hdr.Version > lsgo.MaxVersion {
		return lsgo.Resource{}, fmt.Errorf("LSF version %v is not supported", hdr.Version)
	}

	isCompressed := lsgo.CompressionFlagsToMethod(hdr.CompressionFlags) != lsgo.CMNone && lsgo.CompressionFlagsToMethod(hdr.CompressionFlags) != lsgo.CMInvalid

	pos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF names", "start position", pos)
	if hdr.StringsSizeOnDisk > 0 || hdr.StringsUncompressedSize > 0 {
		uncompressed := lsgo.LimitReadSeeker(r, int64(hdr.StringsSizeOnDisk))

		if isCompressed {
			uncompressed = lsgo.Decompress(uncompressed, int(hdr.StringsUncompressedSize), hdr.CompressionFlags, false)
		}

		// using (var nodesFile = new FileStream("names.bin", FileMode.Create, FileAccess.Write))
		// {
		//     nodesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		names, err = ReadNames(uncompressed)
		if err != nil && err != io.EOF {
			return lsgo.Resource{}, err
		}
	}

	npos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF nodes", "start position", npos)
	if npos != pos+int64(hdr.StringsSizeOnDisk) {
		l.Log("member", "LSF nodes", "msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.StringsSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.StringsSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	if hdr.NodesSizeOnDisk > 0 || hdr.NodesUncompressedSize > 0 {
		uncompressed := lsgo.LimitReadSeeker(r, int64(hdr.NodesSizeOnDisk))
		if isCompressed {
			uncompressed = lsgo.Decompress(uncompressed, int(hdr.NodesUncompressedSize), hdr.CompressionFlags, hdr.Version >= lsgo.VerChunkedCompress)
		}

		// using (var nodesFile = new FileStream("nodes.bin", FileMode.Create, FileAccess.Write))
		// {
		//     nodesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		longNodes := hdr.Version >= lsgo.VerExtendedNodes && hdr.Extended == 1
		nodeInfo, err = readNodeInfo(uncompressed, longNodes)
		if err != nil && err != io.EOF {
			return lsgo.Resource{}, err
		}
	}

	npos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF attributes", "start position", npos)
	if npos != pos+int64(hdr.NodesSizeOnDisk) {
		l.Log("msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.NodesSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.NodesSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	if hdr.AttributesSizeOnDisk > 0 || hdr.AttributesUncompressedSize > 0 {
		var uncompressed io.ReadSeeker = lsgo.LimitReadSeeker(r, int64(hdr.AttributesSizeOnDisk))
		if isCompressed {
			uncompressed = lsgo.Decompress(uncompressed, int(hdr.AttributesUncompressedSize), hdr.CompressionFlags, hdr.Version >= lsgo.VerChunkedCompress)
		}

		// using (var attributesFile = new FileStream("attributes.bin", FileMode.Create, FileAccess.Write))
		// {
		//     attributesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		longAttributes := hdr.Version >= lsgo.VerExtendedNodes && hdr.Extended == 1
		attributeInfo = readAttributeInfo(uncompressed, longAttributes)
	}

	npos, _ = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF values", "start position", npos)
	if npos != pos+int64(hdr.AttributesSizeOnDisk) {
		l.Log("msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.AttributesSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.AttributesSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	var uncompressed io.ReadSeeker = lsgo.LimitReadSeeker(r, int64(hdr.ValuesSizeOnDisk))
	if hdr.ValuesSizeOnDisk > 0 || hdr.ValuesUncompressedSize > 0 {
		if isCompressed {
			uncompressed = lsgo.Decompress(r, int(hdr.ValuesUncompressedSize), hdr.CompressionFlags, hdr.Version >= lsgo.VerChunkedCompress)
		}
	}

	res := lsgo.Resource{}
	valueStart, _ := uncompressed.Seek(0, io.SeekCurrent)
	nodeInstances, err = ReadRegions(uncompressed, valueStart, names, nodeInfo, attributeInfo, hdr.Version, hdr.EngineVersion)
	if err != nil {
		return res, err
	}
	for _, v := range nodeInstances {
		if v.Parent == nil {
			res.Regions = append(res.Regions, v)
		}
	}

	res.Metadata.Major = (hdr.EngineVersion & 0xf0000000) >> 28
	res.Metadata.Minor = (hdr.EngineVersion & 0xf000000) >> 24
	res.Metadata.Revision = (hdr.EngineVersion & 0xff0000) >> 16
	res.Metadata.Build = (hdr.EngineVersion & 0xffff)

	return res, nil
}

func ReadRegions(r io.ReadSeeker, valueStart int64, names [][]string, nodeInfo []NodeInfo, attributeInfo []AttributeInfo, version lsgo.FileVersion, engineVersion uint32) ([]*lsgo.Node, error) {
	NodeInstances := make([]*lsgo.Node, 0, len(nodeInfo))
	for _, nodeInfo := range nodeInfo {
		if nodeInfo.ParentIndex == -1 {
			region, err := ReadNode(r, valueStart, nodeInfo, names, attributeInfo, version, engineVersion)

			region.RegionName = region.Name
			NodeInstances = append(NodeInstances, &region)

			if err != nil {
				return NodeInstances, err
			}
		} else {
			node, err := ReadNode(r, valueStart, nodeInfo, names, attributeInfo, version, engineVersion)

			node.Parent = NodeInstances[nodeInfo.ParentIndex]
			NodeInstances = append(NodeInstances, &node)
			NodeInstances[nodeInfo.ParentIndex].AppendChild(&node)

			if err != nil {
				return NodeInstances, err
			}
		}
	}
	return NodeInstances, nil
}

func ReadNode(r io.ReadSeeker, valueStart int64, ni NodeInfo, names [][]string, attributeInfo []AttributeInfo, version lsgo.FileVersion, engineVersion uint32) (lsgo.Node, error) {
	var (
		node  = lsgo.Node{}
		index = ni.FirstAttributeIndex
		err   error

		l   log.Logger
		pos int64
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "node")
	pos, _ = r.Seek(0, io.SeekCurrent)

	node.Name = names[ni.NameIndex][ni.NameOffset]

	l.Log("member", "name", "read", 0, "start position", pos, "value", node.Name)

	for index != -1 {
		var (
			attribute = attributeInfo[index]
			v         lsgo.NodeAttribute
		)

		if valueStart+int64(attribute.DataOffset) != pos {
			pos, err = r.Seek(valueStart+int64(attribute.DataOffset), io.SeekStart)
			if valueStart+int64(attribute.DataOffset) != pos || err != nil {
				panic("shit")
			}
		}
		v, err = ReadLSFAttribute(r, names[attribute.NameIndex][attribute.NameOffset], attribute.TypeID, attribute.Length, version, engineVersion)
		node.Attributes = append(node.Attributes, v)
		if err != nil {
			return node, err
		}
		index = attribute.NextAttributeIndex
	}
	return node, nil
}

func ReadLSFAttribute(r io.ReadSeeker, name string, dt lsgo.DataType, length uint, version lsgo.FileVersion, engineVersion uint32) (lsgo.NodeAttribute, error) {
	// LSF and LSB serialize the buffer types differently, so specialized
	// code is added to the LSB and LSf serializers, and the common code is
	// available in BinUtils.ReadAttribute()
	var (
		attr = lsgo.NodeAttribute{
			Type: dt,
			Name: name,
		}
		err error

		l   log.Logger
		pos int64
	)
	l = log.With(lsgo.Logger, "component", "LS converter", "file type", "lsf", "part", "attribute")
	pos, _ = r.Seek(0, io.SeekCurrent)

	switch dt {
	case lsgo.DTString, lsgo.DTPath, lsgo.DTFixedString, lsgo.DTLSString, lsgo.DTWString, lsgo.DTLSWString:
		var v string
		v, err = lsgo.ReadCString(r, int(length))
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case lsgo.DTTranslatedString:
		var v lsgo.TranslatedString
		v, err = lsgo.ReadTranslatedString(r, version, engineVersion)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case lsgo.DTTranslatedFSString:
		var v lsgo.TranslatedFSString
		v, err = lsgo.ReadTranslatedFSString(r, version)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	case lsgo.DTScratchBuffer:

		v := make([]byte, length)
		_, err = r.Read(v)
		attr.Value = v

		l.Log("member", name, "read", length, "start position", pos, "value", attr.Value)
		pos += int64(length)

		return attr, err

	default:
		return lsgo.ReadAttribute(r, name, dt, length, l)
	}
}

func init() {
	lsgo.RegisterFormat("lsf", Signature, Read)
}
