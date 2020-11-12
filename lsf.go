package lslib

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	LSFSignature = [4]byte{0x4C, 0x53, 0x4F, 0x46}
	logger       = log.New(os.Stderr, "lslib:", log.LstdFlags|log.Lshortfile)
)

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

func (lsfh *LSFHeader) Read(r io.Reader) error {
	_, err := r.Read(lsfh.Signature[:])
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.Version)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.EngineVersion)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.StringsUncompressedSize)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.StringsSizeOnDisk)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.NodesUncompressedSize)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.NodesSizeOnDisk)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.AttributesUncompressedSize)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.AttributesSizeOnDisk)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.ValuesUncompressedSize)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.ValuesSizeOnDisk)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.CompressionFlags)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.Unknown2)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.Unknown3)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &lsfh.Extended)
	// n, _ := r.Seek(0, io.SeekCurrent)
	// log.Printf("extended flags %#X", lsfh.Extended)
	// log.Printf("current location %v", n)
	if err != nil {
		return err
	}
	// log.Print("Is Compressed: ", lsfh.IsCompressed())
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

func (ne *NodeEntry) Read(r io.Reader) error {
	if ne.Long {
		return ne.readLong(r)
	}
	return ne.readShort(r)
}

func (ne *NodeEntry) readShort(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &ne.NameHashTableIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ne.FirstAttributeIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ne.ParentIndex)
	if err != nil {
		return err
	}
	return nil
}

func (ne *NodeEntry) readLong(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &ne.NameHashTableIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ne.ParentIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ne.NextSiblingIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ne.FirstAttributeIndex)
	if err != nil {
		return err
	}

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

func (ae *AttributeEntry) Read(r io.Reader) error {
	if ae.Long {
		return ae.readLong(r)
	}
	return ae.readShort(r)
}

func (ae *AttributeEntry) readShort(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &ae.NameHashTableIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ae.TypeAndLength)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ae.NodeIndex)
	if err != nil {
		return err
	}
	return nil
}

func (ae *AttributeEntry) readLong(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &ae.NameHashTableIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ae.TypeAndLength)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ae.NextAttributeIndex)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &ae.Offset)
	if err != nil {
		return err
	}

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
func ReadNames(r io.Reader) ([][]string, error) {
	var (
		numHashEntries uint32
		err            error
		names          [][]string
	)
	// n, _ := r.Seek(0, io.SeekCurrent)
	// logger.Print("current location: ", n)
	err = binary.Read(r, binary.LittleEndian, &numHashEntries)
	// logger.Print("names size: ", numHashEntries)
	if err != nil {
		return nil, err
	}
	names = make([][]string, int(numHashEntries))
	for i, _ := range names {

		var numStrings uint16

		err = binary.Read(r, binary.LittleEndian, &numStrings)
		// n, _ = r.Seek(0, io.SeekCurrent)
		// logger.Print("current location: ", n, " name count: ", numStrings)

		var hash = make([]string, int(numStrings))
		for x, _ := range hash {
			var (
				nameLen uint16
				name    []byte
			)
			err = binary.Read(r, binary.LittleEndian, &nameLen)
			if err != nil {
				return nil, err
			}
			name = make([]byte, nameLen)

			_, err = r.Read(name)
			if err != nil {
				return nil, err
			}
			hash[x] = string(name)

		}
		names[i] = hash

	}
	return names, nil
}

func readNodeInfo(r io.Reader, longNodes bool) ([]NodeInfo, error) {
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
func readAttributeInfo(r io.Reader, long bool) []AttributeInfo {
	// var rawAttributes = new List<AttributeEntryV2>();

	var (
		prevAttributeRefs []int
		dataOffset        uint = 0
		index                  = 0
		nextAttrIndex     int  = -1
		attributes        []AttributeInfo
		err               error
	)
	for err == nil {
		attribute := &AttributeEntry{Long: long}
		err = attribute.Read(r)
		// pretty.Log(err, attribute)
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
			nodeIndex := int(attribute.NodeIndex + 1)
			if len(prevAttributeRefs) > int(nodeIndex) {
				if prevAttributeRefs[nodeIndex] != -1 {
					attributes[prevAttributeRefs[nodeIndex]].NextAttributeIndex = index
				}

				prevAttributeRefs[nodeIndex] = index
			} else {
				for len(prevAttributeRefs) < nodeIndex {
					prevAttributeRefs = append(prevAttributeRefs, -1)
				}

				prevAttributeRefs = append(prevAttributeRefs, index)
			}

			// rawAttributes.Add(attribute);
			dataOffset += uint(resolved.Length)
		}

		attributes = append(attributes, resolved)
		index++
	}
	return attributes[:len(attributes)-1]
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

func ReadLSF(r io.Reader) (Resource, error) {
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
	// s, ok := r.(*bufio.Reader)
	// if ok {
	// 	lsfr.data = s
	// } else {
	// 	lsfr.data = bufio.NewReader(r)
	// }
	hdr := &LSFHeader{}
	err = hdr.Read(r)
	// logger.Print("Signature: ", string(hdr.Signature[:]))
	// pretty.Log(hdr)
	if err != nil || (hdr.Signature != LSFSignature) {
		return Resource{}, fmt.Errorf("Invalid LSF signature; expected %v, got %v", LSFSignature, hdr.Signature)
	}

	if hdr.Version < VerInitial || hdr.Version > MaxVersion {
		return Resource{}, fmt.Errorf("LSF version %v is not supported", hdr.Version)
	}

	// Names = new List<List<String>>();
	isCompressed := CompressionFlagsToMethod(hdr.CompressionFlags) != CMNone && CompressionFlagsToMethod(hdr.CompressionFlags) != CMInvalid
	// logger.Printf("Names compressed: %v; Compression Flags: %v", isCompressed, CompressionFlagsToMethod(hdr.CompressionFlags))
	if hdr.StringsSizeOnDisk > 0 || hdr.StringsUncompressedSize > 0 {
		var (
			uncompressed = io.LimitReader(r, int64(hdr.StringsSizeOnDisk))
		)

		if isCompressed {
			uncompressed = Decompress(uncompressed, hdr.CompressionFlags, false)
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

	// Nodes = new List<NodeInfo>();
	if hdr.NodesSizeOnDisk > 0 || hdr.NodesUncompressedSize > 0 {
		var (
			uncompressed = io.LimitReader(r, int64(hdr.NodesSizeOnDisk))
		)
		if isCompressed {
			uncompressed = Decompress(uncompressed, hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
		}

		// using (var nodesFile = new FileStream("nodes.bin", FileMode.Create, FileAccess.Write))
		// {
		//     nodesFile.Write(uncompressed, 0, uncompressed.Length);
		// }

		longNodes := hdr.Version >= VerExtendedNodes && hdr.Extended == 1
		nodeInfo, err = readNodeInfo(uncompressed, longNodes)
		// logger.Printf("region 1 name: %v", names[nodeInfo[0].NameIndex])
		// pretty.Log(nodeInfo)
		if err != nil && err != io.EOF {
			return Resource{}, err
		}
	}

	// Attributes = new List<AttributeInfo>();
	if hdr.AttributesSizeOnDisk > 0 || hdr.AttributesUncompressedSize > 0 {
		var (
			uncompressed io.Reader = io.LimitReader(r, int64(hdr.AttributesSizeOnDisk))
		)
		if isCompressed {
			uncompressed = Decompress(uncompressed, hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
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

	var (
		uncompressed io.Reader = io.LimitReader(r, int64(hdr.ValuesSizeOnDisk))
	)
	if hdr.ValuesSizeOnDisk > 0 || hdr.ValuesUncompressedSize > 0 {
		if isCompressed {
			uncompressed = Decompress(r, hdr.CompressionFlags, hdr.Version >= VerChunkedCompress)
		}

		// using (var valuesFile = new FileStream("values.bin", FileMode.Create, FileAccess.Write))
		// {
		//     valuesFile.Write(uncompressed, 0, uncompressed.Length);
		// }
	}

	res := Resource{}
	// n, _ := r.Seek(0, io.SeekCurrent)
	nodeInstances, err = ReadRegions(uncompressed, names, nodeInfo, attributeInfo, hdr.Version)
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

func ReadRegions(r io.Reader, names [][]string, nodeInfo []NodeInfo, attributeInfo []AttributeInfo, Version FileVersion) ([]*Node, error) {
	NodeInstances := make([]*Node, 0, len(nodeInfo))
	for _, nodeInfo := range nodeInfo {
		if nodeInfo.ParentIndex == -1 {
			region, err := ReadNode(r, nodeInfo, names, attributeInfo, Version)

			// pretty.Log(err, region)

			region.RegionName = region.Name
			NodeInstances = append(NodeInstances, &region)

			if err != nil {
				return NodeInstances, err
			}
		} else {
			node, err := ReadNode(r, nodeInfo, names, attributeInfo, Version)

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

func ReadNode(r io.Reader, ni NodeInfo, names [][]string, attributeInfo []AttributeInfo, Version FileVersion) (Node, error) {
	var (
		node  = Node{}
		index = ni.FirstAttributeIndex
		err   error
	)
	// pretty.Log(ni)
	node.Name = names[ni.NameIndex][ni.NameOffset]
	logger.Printf("reading node %s", names[ni.NameIndex][ni.NameOffset])

	for index != -1 {
		var (
			attribute = attributeInfo[index]
			v         NodeAttribute
		)
		// n, _ := r.Seek(0, io.SeekCurrent)
		// logger.Printf("seeking to %v now at %v", attribute.DataOffset, n)
		v, err = ReadLSFAttribute(r, names[attribute.NameIndex][attribute.NameOffset], attribute.TypeId, attribute.Length, Version)
		node.Attributes = append(node.Attributes, v)
		if err != nil {
			return node, err
		}
		index = attribute.NextAttributeIndex

		// Console.WriteLine(String.Format("    {0:X}: {1} ({2})", attribute.DataOffset, names[attribute.NameIndex][attribute.NameOffset], value));
	}
	return node, nil
}

func ReadLSFAttribute(r io.Reader, name string, DT DataType, length uint, Version FileVersion) (NodeAttribute, error) {
	// LSF and LSB serialize the buffer types differently, so specialized
	// code is added to the LSB and LSf serializers, and the common code is
	// available in BinUtils.ReadAttribute()
	var (
		attr = NodeAttribute{
			Type: DT,
			Name: name,
		}
		err error
	)
	logger.Printf("reading attribute '%v' type %v of length %v", name, DT, length)
	switch DT {
	case DT_String, DT_Path, DT_FixedString, DT_LSString, DT_WString, DT_LSWString:
		var v string
		v, err = ReadCString(r, int(length))
		attr.Value = v
		return attr, err

	case DT_TranslatedString:
		var str TranslatedString

		if Version >= VerBG3 {
			// logger.Println("decoding bg3 data")
			var version uint16
			/*err =*/ binary.Read(r, binary.LittleEndian, &version)
			str.Version = version
		} else {
			str.Version = 0

			var (
				vlength int32
				v       []byte
			)

			/*err =*/
			binary.Read(r, binary.LittleEndian, &vlength)
			v = make([]byte, length)
			/*err =*/ r.Read(v)
			str.Value = string(v)
		}

		var handleLength int32
		/*err =*/ binary.Read(r, binary.LittleEndian, &handleLength)
		str.Handle, err = ReadCString(r, int(handleLength))
		// logger.Printf("string %s; %v", str.Handle, err)

		attr.Value = str
		return attr, err

	case DT_TranslatedFSString:
		var v TranslatedFSString
		v, err = ReadTranslatedFSString(r, Version)
		attr.Value = v
		return attr, err

	case DT_ScratchBuffer:

		v := make([]byte, length)
		_, err = r.Read(v)
		attr.Value = v
		return attr, err

	default:
		return ReadAttribute(r, name, DT)
	}
}

func ReadTranslatedFSString(r io.Reader, Version FileVersion) (TranslatedFSString, error) {
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
