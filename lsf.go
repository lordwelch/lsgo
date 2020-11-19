package lslib

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
)

var (
	LSFSignature            = [4]byte{0x4C, 0x53, 0x4F, 0x46}
	Logger       log.Logger = log.NewNopLogger()
)

// NewFilter allows filtering of l
func NewFilter(f map[string][]string, l log.Logger) log.Logger {
	return filter{
		filter: f,
		next:   l,
	}
}

type filter struct {
	next   log.Logger
	filter map[string][]string
}

func (f filter) Log(keyvals ...interface{}) error {
	var allowed = true // allow everything
	for i := 0; i < len(keyvals)-1; i += 2 {
		if v, ok := keyvals[i].(string); ok { // key
			if fil, ok := f.filter[v]; ok { // key has a filter
				if v, ok = keyvals[i+1].(string); ok { // value is a string
					allowed = false // this key has a filter deny everything except what the filter allows
					for _, fi := range fil {
						if strings.Contains(v, fi) {
							allowed = true
						}
					}
				}
			}
		}
	}
	if allowed {
		return f.next.Log(keyvals...)
	}
	return nil
}

type LSFHeader struct {
	/// summary
	/// LSOF file signature
	/// /summary
	Signature [4]byte

	/// summary
	/// Version of the LSOF file D:OS EE is version 1/2, D:OS 2 is version 3
	/// /summary
	Version FileVersion
	/// summary
	/// Possibly version number? (major, minor, rev, build)
	/// /summary
	EngineVersion uint32
	/// summary
	/// Total uncompressed size of the string hash table
	/// /summary
	StringsUncompressedSize uint32
	/// summary
	/// Compressed size of the string hash table
	/// /summary
	StringsSizeOnDisk uint32
	/// summary
	/// Total uncompressed size of the node list
	/// /summary
	NodesUncompressedSize uint32
	/// summary
	/// Compressed size of the node list
	/// /summary
	NodesSizeOnDisk uint32
	/// summary
	/// Total uncompressed size of the attribute list
	/// /summary
	AttributesUncompressedSize uint32
	/// summary
	/// Compressed size of the attribute list
	/// /summary
	AttributesSizeOnDisk uint32
	/// summary
	/// Total uncompressed size of the raw value buffer
	/// /summary
	ValuesUncompressedSize uint32
	/// summary
	/// Compressed size of the raw value buffer
	/// /summary
	ValuesSizeOnDisk uint32
	/// summary
	/// Compression method and level used for the string, node, attribute and value buffers.
	/// Uses the same format as packages (see BinUtils.MakeCompressionFlags)
	/// /summary
	CompressionFlags byte
	/// summary
	/// Possibly unused, always 0
	/// /summary
	Unknown2 byte
	Unknown3 uint16
	/// summary
	/// Extended node/attribute format indicator, 0 for V2, 0/1 for V3
	/// /summary
	Extended uint32
}

func (lsfh *LSFHeader) Read(r io.ReadSeeker) error {
	var (
		l   log.Logger
		pos int64
		n   int
		err error
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "header")
	pos, err = r.Seek(0, io.SeekCurrent)
	n, err = r.Read(lsfh.Signature[:])
	if err != nil {
		return err
	}
	l.Log("member", "Signature", "read", n, "start position", pos, "value", string(lsfh.Signature[:]))
	pos += int64(n)
	err = binary.Read(r, binary.LittleEndian, &lsfh.Version)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Version", "read", n, "start position", pos, "value", lsfh.Version)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.EngineVersion)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "EngineVersion", "read", n, "start position", pos, "value", fmt.Sprintf("%d.%d.%d.%d", (lsfh.EngineVersion&0xf0000000)>>28, (lsfh.EngineVersion&0xf000000)>>24, (lsfh.EngineVersion&0xff0000)>>16, (lsfh.EngineVersion&0xffff)))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.StringsUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "StringsUncompressedSize", "read", n, "start position", pos, "value", lsfh.StringsUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.StringsSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "StringsSizeOnDisk", "read", n, "start position", pos, "value", lsfh.StringsSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.NodesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NodesUncompressedSize", "read", n, "start position", pos, "value", lsfh.NodesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.NodesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "NodesSizeOnDisk", "read", n, "start position", pos, "value", lsfh.NodesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.AttributesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "AttributesUncompressedSize", "read", n, "start position", pos, "value", lsfh.AttributesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.AttributesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "AttributesSizeOnDisk", "read", n, "start position", pos, "value", lsfh.AttributesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.ValuesUncompressedSize)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "ValuesUncompressedSize", "read", n, "start position", pos, "value", lsfh.ValuesUncompressedSize)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.ValuesSizeOnDisk)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "ValuesSizeOnDisk", "read", n, "start position", pos, "value", lsfh.ValuesSizeOnDisk)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.CompressionFlags)
	n = 1
	if err != nil {
		return err
	}
	l.Log("member", "CompressionFlags", "read", n, "start position", pos, "value", lsfh.CompressionFlags)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.Unknown2)
	n = 1
	if err != nil {
		return err
	}
	l.Log("member", "Unknown2", "read", n, "start position", pos, "value", lsfh.Unknown2)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.Unknown3)
	n = 2
	if err != nil {
		return err
	}
	l.Log("member", "Unknown3", "read", n, "start position", pos, "value", lsfh.Unknown3)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &lsfh.Extended)
	n = 4
	if err != nil {
		return err
	}
	l.Log("member", "Extended", "read", n, "start position", pos, "value", lsfh.Extended)
	pos += int64(n)

	if !lsfh.IsCompressed() {
		lsfh.NodesSizeOnDisk = lsfh.NodesUncompressedSize
		lsfh.AttributesSizeOnDisk = lsfh.AttributesUncompressedSize
		lsfh.StringsSizeOnDisk = lsfh.StringsUncompressedSize
		lsfh.ValuesSizeOnDisk = lsfh.ValuesUncompressedSize
	}
	return nil
}

func (lsfh LSFHeader) IsCompressed() bool {
	return CompressionFlagsToMethod(lsfh.CompressionFlags) != CMNone && CompressionFlagsToMethod(lsfh.CompressionFlags) != CMInvalid
}

type NodeEntry struct {
	Long bool

	/// summary
	/// Name of this node
	/// (16-bit MSB: index into name hash table, 16-bit LSB: offset in hash chain)
	/// /summary
	NameHashTableIndex uint32

	/// summary
	/// Index of the first attribute of this node
	/// (-1: node has no attributes)
	/// /summary
	FirstAttributeIndex int32

	/// summary
	/// Index of the parent node
	/// (-1: this node is a root region)
	/// /summary
	ParentIndex int32

	/// summary
	/// Index of the next sibling of this node
	/// (-1: this is the last node)
	/// /summary
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "short node")
	pos, err = r.Seek(0, io.SeekCurrent)
	err = binary.Read(r, binary.LittleEndian, &ne.NameHashTableIndex)
	n = 4
	if err != nil {
		// logger.Println(err, "ne.NameHashTableIndex", ne.NameHashTableIndex)
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", strconv.Itoa(ne.NameIndex())+" "+strconv.Itoa(ne.NameOffset()))
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.FirstAttributeIndex)
	n = 4
	if err != nil {
		// logger.Println(err, "ne.FirstAttributeIndex", ne.FirstAttributeIndex)
		return err
	}
	l.Log("member", "NameHashTableIndex", "read", n, "start position", pos, "value", ne.FirstAttributeIndex)
	pos += int64(n)

	err = binary.Read(r, binary.LittleEndian, &ne.ParentIndex)
	n = 4
	if err != nil {
		// logger.Println(err, "ne.ParentIndex", ne.ParentIndex)
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "long node")
	pos, err = r.Seek(0, io.SeekCurrent)
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

/// summary
/// Processed node information for a node in the LSF file
/// /summary
type NodeInfo struct {
	/// summary
	/// Index of the parent node
	/// (-1: this node is a root region)
	/// /summary
	ParentIndex int

	/// summary
	/// Index into name hash table
	/// /summary
	NameIndex int

	/// summary
	/// Offset in hash chain
	/// /summary
	NameOffset int

	/// summary
	/// Index of the first attribute of this node
	/// (-1: node has no attributes)
	/// /summary
	FirstAttributeIndex int
}

/// summary
/// attribute extension in the LSF file
/// /summary
type AttributeEntry struct {
	Long bool
	/// summary
	/// Name of this attribute
	/// (16-bit MSB: index into name hash table, 16-bit LSB: offset in hash chain)
	/// /summary
	NameHashTableIndex uint32

	/// summary
	/// 6-bit LSB: Type of this attribute (see NodeAttribute.DataType)
	/// 26-bit MSB: Length of this attribute
	/// /summary
	TypeAndLength uint32

	/// summary
	/// Index of the node that this attribute belongs to
	/// Note: These indexes are assigned seemingly arbitrarily, and are not neccessarily indices into the node list
	/// /summary
	NodeIndex int32

	/// summary
	/// Index of the node that this attribute belongs to
	/// Note: These indexes are assigned seemingly arbitrarily, and are not neccessarily indices into the node list
	/// /summary
	NextAttributeIndex int32

	/// <summary>
	/// Absolute position of attribute value in the value stream
	/// </summary>
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "short attribute")
	pos, err = r.Seek(0, io.SeekCurrent)

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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "long attribute")
	pos, err = r.Seek(0, io.SeekCurrent)

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

/// summary
/// Index into name hash table
/// /summary
func (ae AttributeEntry) NameIndex() int {
	return int(ae.NameHashTableIndex >> 16)
}

/// summary
/// Offset in hash chain
/// /summary
func (ae AttributeEntry) NameOffset() int {
	return int(ae.NameHashTableIndex & 0xffff)
}

/// summary
/// Type of this attribute (see NodeAttribute.DataType)
/// /summary
func (ae AttributeEntry) TypeID() DataType {
	return DataType(ae.TypeAndLength & 0x3f)
}

/// summary
/// Length of this attribute
/// /summary
func (ae AttributeEntry) Len() int {
	return int(ae.TypeAndLength >> 6)
}

type AttributeInfo struct {
	V2 bool

	/// summary
	/// Index into name hash table
	/// /summary
	NameIndex int
	/// summary
	/// Offset in hash chain
	/// /summary
	NameOffset int
	/// summary
	/// Type of this attribute (see NodeAttribute.DataType)
	/// /summary
	TypeId DataType
	/// summary
	/// Length of this attribute
	/// /summary
	Length uint
	/// summary
	/// Absolute position of attribute data in the values section
	/// /summary
	DataOffset uint
	/// summary
	/// Index of the next attribute in this node
	/// (-1: this is the last attribute)
	/// /summary
	NextAttributeIndex int
}

type LSFReader struct {
	data *bufio.Reader
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "names")
	pos, err = r.Seek(0, io.SeekCurrent)

	err = binary.Read(r, binary.LittleEndian, &numHashEntries)
	n = 4
	if err != nil {
		return nil, err
	}
	l.Log("member", "numHashEntries", "read", n, "start position", pos, "value", numHashEntries)
	pos += int64(n)

	names = make([][]string, int(numHashEntries))
	for i, _ := range names {

		var numStrings uint16

		err = binary.Read(r, binary.LittleEndian, &numStrings)
		n = 4
		l.Log("member", "numStrings", "read", n, "start position", pos, "value", numStrings)
		pos += int64(n)

		var hash = make([]string, int(numStrings))
		for x, _ := range hash {
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
	// Console.WriteLine(" ----- DUMP OF NODE TABLE -----");
	var (
		nodes []NodeInfo
		err   error
	)
	index := 0

	for err == nil {
		var node NodeInfo
		// var pos = lsfr.Position;

		item := &NodeEntry{Long: longNodes}
		err = item.Read(r)

		node.FirstAttributeIndex = int(item.FirstAttributeIndex)
		node.NameIndex = item.NameIndex()
		node.NameOffset = item.NameOffset()
		node.ParentIndex = int(item.ParentIndex)
		// Console.WriteLine(String.Format(
		//     "{0}: {1} @ {2:X} (parent {3}, firstAttribute {4})",
		//     index, Names[node.NameIndex][node.NameOffset], pos, node.ParentIndex,
		//     node.FirstAttributeIndex
		// ));

		nodes = append(nodes, node)
		index++
	}
	return nodes[:len(nodes)-1], err
}

/// <summary>
/// Reads the attribute headers for the LSOF resource
/// </summary>
/// <param name="s">Stream to read the attribute headers from</param>
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

		// pretty.Log( attribute)
		if long {
			dataOffset = uint(attribute.Offset)
			nextAttrIndex = int(attribute.NextAttributeIndex)
		}

		resolved := AttributeInfo{
			NameIndex:          attribute.NameIndex(),
			NameOffset:         attribute.NameOffset(),
			TypeId:             attribute.TypeID(),
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
			dataOffset += uint(resolved.Length)

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

type HeaderError struct {
	Expected []byte
	Got      []byte
}

func (he HeaderError) Error() string {
	return fmt.Sprintf("Invalid LSF signature; expected %v, got %v", he.Expected, he.Got)
}

func ReadLSF(r io.ReadSeeker) (Resource, error) {
	var (
		err error
		/// summary
		/// Static string hash map
		/// /summary
		names [][]string
		/// summary
		/// Preprocessed list of nodes (structures)
		/// /summary
		nodeInfo []NodeInfo
		/// summary
		/// Preprocessed list of node attributes
		/// /summary
		attributeInfo []AttributeInfo
		/// summary
		/// Node instances
		/// /summary
		nodeInstances []*Node
	)
	var (
		l         log.Logger
		pos, npos int64
		// n   int
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "file")
	pos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF header", "start position", pos)

	hdr := &LSFHeader{}
	err = hdr.Read(r)
	if err != nil || (hdr.Signature != LSFSignature) {
		return Resource{}, HeaderError{LSFSignature[:], hdr.Signature[:]}
	}

	if hdr.Version < VerInitial || hdr.Version > MaxVersion {
		return Resource{}, fmt.Errorf("LSF version %v is not supported", hdr.Version)
	}

	isCompressed := CompressionFlagsToMethod(hdr.CompressionFlags) != CMNone && CompressionFlagsToMethod(hdr.CompressionFlags) != CMInvalid

	pos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF names", "start position", pos)
	if hdr.StringsSizeOnDisk > 0 || hdr.StringsUncompressedSize > 0 {
		var (
			uncompressed = LimitReadSeeker(r, int64(hdr.StringsSizeOnDisk))
		)

		if isCompressed {
			uncompressed = Decompress(uncompressed, int(hdr.StringsUncompressedSize), hdr.CompressionFlags, false)
		}

		// using (var nodesFile = new FileStream("names.bin", FileMode.Create, FileAccess.Write))
		// {
		//     nodesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		names, err = ReadNames(uncompressed)
		// pretty.Log(len(names), names)
		if err != nil && err != io.EOF {
			return Resource{}, err
		}
	}

	npos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF nodes", "start position", npos)
	if npos != pos+int64(hdr.StringsSizeOnDisk) {
		l.Log("member", "LSF nodes", "msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.StringsSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.StringsSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	if hdr.NodesSizeOnDisk > 0 || hdr.NodesUncompressedSize > 0 {
		var (
			uncompressed = LimitReadSeeker(r, int64(hdr.NodesSizeOnDisk))
		)
		if isCompressed {
			uncompressed = Decompress(uncompressed, int(hdr.NodesUncompressedSize), hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
		}

		// using (var nodesFile = new FileStream("nodes.bin", FileMode.Create, FileAccess.Write))
		// {
		//     nodesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		longNodes := hdr.Version >= VerExtendedNodes && hdr.Extended == 1
		nodeInfo, err = readNodeInfo(uncompressed, longNodes)
		// pretty.Log(err, nodeInfo)
		// logger.Printf("region 1 name: %v", names[nodeInfo[0].NameIndex])
		if err != nil && err != io.EOF {
			return Resource{}, err
		}
	}

	npos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF attributes", "start position", npos)
	if npos != pos+int64(hdr.NodesSizeOnDisk) {
		l.Log("msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.NodesSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.NodesSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	if hdr.AttributesSizeOnDisk > 0 || hdr.AttributesUncompressedSize > 0 {
		var (
			uncompressed io.ReadSeeker = LimitReadSeeker(r, int64(hdr.AttributesSizeOnDisk))
		)
		if isCompressed {
			uncompressed = Decompress(uncompressed, int(hdr.AttributesUncompressedSize), hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
		}

		// using (var attributesFile = new FileStream("attributes.bin", FileMode.Create, FileAccess.Write))
		// {
		//     attributesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		longAttributes := hdr.Version >= VerExtendedNodes && hdr.Extended == 1
		attributeInfo = readAttributeInfo(uncompressed, longAttributes)
		// logger.Printf("attribute 1 name: %v", names[attributeInfo[0].NameIndex])
		// pretty.Log(attributeInfo)
	}

	npos, err = r.Seek(0, io.SeekCurrent)
	l.Log("member", "LSF values", "start position", npos)
	if npos != pos+int64(hdr.AttributesSizeOnDisk) {
		l.Log("msg", "seeking to correct offset", "current", npos, "wanted", pos+int64(hdr.AttributesSizeOnDisk))
		pos, _ = r.Seek((pos+int64(hdr.AttributesSizeOnDisk))-npos, io.SeekCurrent)
	} else {
		pos = npos
	}
	var (
		uncompressed io.ReadSeeker = LimitReadSeeker(r, int64(hdr.ValuesSizeOnDisk))
	)
	if hdr.ValuesSizeOnDisk > 0 || hdr.ValuesUncompressedSize > 0 {
		if isCompressed {
			uncompressed = Decompress(r, int(hdr.ValuesUncompressedSize), hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
		}
	}

	res := Resource{}
	valueStart, _ = uncompressed.Seek(0, io.SeekCurrent)
	nodeInstances, err = ReadRegions(uncompressed, names, nodeInfo, attributeInfo, hdr.Version, hdr.EngineVersion)
	if err != nil {
		return res, err
	}
	for _, v := range nodeInstances {
		if v.Parent == nil {
			res.Regions = append(res.Regions, v)
		}
	}

	res.Metadata.MajorVersion = (hdr.EngineVersion & 0xf0000000) >> 28
	res.Metadata.MinorVersion = (hdr.EngineVersion & 0xf000000) >> 24
	res.Metadata.Revision = (hdr.EngineVersion & 0xff0000) >> 16
	res.Metadata.BuildNumber = (hdr.EngineVersion & 0xffff)

	// pretty.Log(res)
	return res, nil

}

var valueStart int64

func ReadRegions(r io.ReadSeeker, names [][]string, nodeInfo []NodeInfo, attributeInfo []AttributeInfo, Version FileVersion, EngineVersion uint32) ([]*Node, error) {
	NodeInstances := make([]*Node, 0, len(nodeInfo))
	for _, nodeInfo := range nodeInfo {
		if nodeInfo.ParentIndex == -1 {
			region, err := ReadNode(r, nodeInfo, names, attributeInfo, Version, EngineVersion)

			// pretty.Log(err, region)

			region.RegionName = region.Name
			NodeInstances = append(NodeInstances, &region)

			if err != nil {
				return NodeInstances, err
			}
		} else {
			node, err := ReadNode(r, nodeInfo, names, attributeInfo, Version, EngineVersion)

			// pretty.Log(err, node)

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

func ReadNode(r io.ReadSeeker, ni NodeInfo, names [][]string, attributeInfo []AttributeInfo, Version FileVersion, EngineVersion uint32) (Node, error) {
	var (
		node  = Node{}
		index = ni.FirstAttributeIndex
		err   error

		l   log.Logger
		pos int64
	)
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "node")
	pos, err = r.Seek(0, io.SeekCurrent)

	node.Name = names[ni.NameIndex][ni.NameOffset]

	l.Log("member", "name", "read", 0, "start position", pos, "value", node.Name)

	for index != -1 {
		var (
			attribute = attributeInfo[index]
			v         NodeAttribute
		)

		if valueStart+int64(attribute.DataOffset) != pos {
			pos, err = r.Seek(valueStart+int64(attribute.DataOffset), io.SeekStart)
			if valueStart+int64(attribute.DataOffset) != pos || err != nil {
				panic("shit")
			}
		}
		v, err = ReadLSFAttribute(r, names[attribute.NameIndex][attribute.NameOffset], attribute.TypeId, attribute.Length, Version, EngineVersion)
		node.Attributes = append(node.Attributes, v)
		if err != nil {
			return node, err
		}
		index = attribute.NextAttributeIndex

		// Console.WriteLine(String.Format("    {0:X}: {1} ({2})", attribute.DataOffset, names[attribute.NameIndex][attribute.NameOffset], value));
	}
	return node, nil
}

func ReadLSFAttribute(r io.ReadSeeker, name string, DT DataType, length uint, Version FileVersion, EngineVersion uint32) (NodeAttribute, error) {
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
	l = log.With(Logger, "component", "LS converter", "file type", "lsf", "part", "attribute")
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

func ReadTranslatedString(r io.ReadSeeker, Version FileVersion, EngineVersion uint32) (TranslatedString, error) {
	var (
		str TranslatedString
		err error
	)

	if Version >= VerBG3 || EngineVersion == 0x4000001d {
		// logger.Println("decoding bg3 data")
		var version uint16
		/*err =*/ binary.Read(r, binary.LittleEndian, &version)
		str.Version = version
	} else {
		str.Version = 0

		var (
			vlength int32
			v       []byte
			// n       int
		)

		err = binary.Read(r, binary.LittleEndian, &vlength)
		if err != nil {
			return str, err
		}
		v = make([]byte, vlength)
		_, err = r.Read(v)
		if err != nil {
			return str, err
		}
		// logger.Printf("value length: %d value: %s read length: %d len of v: %d", vlength, v, n, len(v))
		str.Value = string(v)
	}

	var handleLength int32
	err = binary.Read(r, binary.LittleEndian, &handleLength)
	if err != nil {
		return str, err
	}
	str.Handle, err = ReadCString(r, int(handleLength))
	if err != nil {
		return str, err
	}
	// logger.Printf("handle %s; %v", str.Handle, err)
	return str, nil
}

func ReadTranslatedFSString(r io.ReadSeeker, Version FileVersion) (TranslatedFSString, error) {
	var (
		str = TranslatedFSString{}
		err error
	)

	if Version >= VerBG3 {
		var version uint16
		/*err =*/ binary.Read(r, binary.LittleEndian, &version)
		str.Version = version
	} else {
		str.Version = 0

		var (
			length int32
		)

		/*err =*/
		binary.Read(r, binary.LittleEndian, &length)
		str.Value, err = ReadCString(r, int(length))
		if err != nil {
			return str, err
		}
	}

	var handleLength int32
	err = binary.Read(r, binary.LittleEndian, &handleLength)
	if err != nil {
		return str, err
	}
	str.Handle, err = ReadCString(r, int(handleLength))
	if err != nil {
		return str, err
	}

	var arguments int32
	err = binary.Read(r, binary.LittleEndian, &arguments)
	if err != nil {
		return str, err
	}
	str.Arguments = make([]TranslatedFSStringArgument, 0, arguments)
	for i := 0; i < int(arguments); i++ {
		arg := TranslatedFSStringArgument{}

		var argKeyLength int32
		err = binary.Read(r, binary.LittleEndian, &argKeyLength)
		if err != nil {
			return str, err
		}
		arg.Key, err = ReadCString(r, int(argKeyLength))
		if err != nil {
			return str, err
		}

		arg.String, err = ReadTranslatedFSString(r, Version)
		if err != nil {
			return str, err
		}

		var argValueLength int32
		err = binary.Read(r, binary.LittleEndian, &argValueLength)
		if err != nil {
			return str, err
		}
		arg.Value, err = ReadCString(r, int(argValueLength))
		if err != nil {
			return str, err
		}

		str.Arguments = append(str.Arguments, arg)
	}

	return str, nil
}
