package lsgo

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/mat"
)

// XMLMarshaler has a pointer to start in order to append multiple attributes to the xml element
type XMLMarshaler interface {
	MarshalXML(e *xml.Encoder, start *xml.StartElement) error
}

type TranslatedString struct {
	Version uint16
	Value   string
	Handle  string
}

func (ts TranslatedString) MarshalXML(e *xml.Encoder, start *xml.StartElement) error {
	start.Attr = append(start.Attr,
		xml.Attr{
			Name:  xml.Name{Local: "handle"},
			Value: ts.Handle,
		},
		xml.Attr{
			Name:  xml.Name{Local: "version"},
			Value: strconv.Itoa(int(ts.Version)),
		},
	)
	return nil
}

type TranslatedFSStringArgument struct {
	String TranslatedFSString
	Key    string
	Value  string
}

type TranslatedFSString struct {
	TranslatedString
	Arguments []TranslatedFSStringArgument
}

// func (tfs TranslatedFSString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	start.Attr = append(start.Attr,
// 		xml.Attr{
// 			Name:  xml.Name{Local: "version"},
// 			Value: strconv.Itoa(int(tfs.Version)),
// 		},
// 		xml.Attr{
// 			Name:  xml.Name{Local: "handle"},
// 			Value: tfs.Handle,
// 		},
// 		xml.Attr{
// 			Name:  xml.Name{Local: "value"},
// 			Value: ts.Value,
// 		},
// 	)
// 	return nil
// }

type Ivec []int

func (i Ivec) String() string {
	b := &strings.Builder{}
	for _, v := range i {
		b.WriteString(" ")
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()[1:]
}

type Vec []float64

type Mat mat.Dense

func (m Mat) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var (
		M = mat.Dense(m)
		v []float64
	)
	rows, cols := M.Dims()
	if rows == cols {
		start.Name.Local = "mat" + strconv.Itoa(rows)
	} else {
		start.Name.Local = "mat" + strconv.Itoa(rows) + "x" + strconv.Itoa(cols)
	}
	e.EncodeToken(start)
	for i := 0; i < rows; i++ {
		v = M.RawRowView(i)
		n := Vec(v)
		e.Encode(n)
	}
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func (v Vec) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var name xml.Name
	for i := 0; i < len(v); i++ {
		switch i {
		case 0:
			name.Local = "x"
			// start.Name = "float1"
		case 1:
			name.Local = "y"
			start.Name.Local = "float2"
		case 2:
			name.Local = "z"
			start.Name.Local = "float3"
		case 3:
			name.Local = "w"
			start.Name.Local = "float4"

		default:
			return ErrVectorTooBig
		}
		start.Attr = append(start.Attr, xml.Attr{
			Name:  name,
			Value: strconv.FormatFloat(v[i], 'f', -1, 32),
		})
	}
	e.EncodeToken(start)
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

type DataType int

const (
	DTNone DataType = iota
	DTByte
	DTShort
	DTUShort
	DTInt
	DTUInt
	DTFloat
	DTDouble
	DTIVec2
	DTIVec3
	DTIVec4
	DTVec2
	DTVec3
	DTVec4
	DTMat2
	DTMat3
	DTMat3x4
	DTMat4x3
	DTMat4
	DTBool
	DTString
	DTPath
	DTFixedString
	DTLSString
	DTULongLong
	DTScratchBuffer
	// Seems to be unused?
	DTLong
	DTInt8
	DTTranslatedString
	DTWString
	DTLSWString
	DTUUID
	DTInt64
	DTTranslatedFSString
	// Last supported datatype, always keep this one at the end
	DTMax = iota - 1
)

func (dt *DataType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Value: dt.String(),
		Name:  name,
	}, nil
}

func (dt DataType) String() string {
	switch dt {
	case DTNone:
		return "None"
	case DTByte:
		return "uint8"
	case DTShort:
		return "int16"
	case DTUShort:
		return "uint16"
	case DTInt:
		return "int32"
	case DTUInt:
		return "uint32"
	case DTFloat:
		return "float"
	case DTDouble:
		return "double"
	case DTIVec2:
		return "ivec2"
	case DTIVec3:
		return "ivec3"
	case DTIVec4:
		return "ivec4"
	case DTVec2:
		return "fvec2"
	case DTVec3:
		return "fvec3"
	case DTVec4:
		return "fvec4"
	case DTMat2:
		return "mat2x2"
	case DTMat3:
		return "mat3x3"
	case DTMat3x4:
		return "mat3x4"
	case DTMat4x3:
		return "mat4x3"
	case DTMat4:
		return "mat4x4"
	case DTBool:
		return "bool"
	case DTString:
		return "string"
	case DTPath:
		return "path"
	case DTFixedString:
		return "FixedString"
	case DTLSString:
		return "LSString"
	case DTULongLong:
		return "uint64"
	case DTScratchBuffer:
		return "ScratchBuffer"
	case DTLong:
		return "old_int64"
	case DTInt8:
		return "int8"
	case DTTranslatedString:
		return "TranslatedString"
	case DTWString:
		return "WString"
	case DTLSWString:
		return "LSWString"
	case DTUUID:
		return "guid"
	case DTInt64:
		return "int64"
	case DTTranslatedFSString:
		return "TranslatedFSString"
	}
	return ""
}

type NodeAttribute struct {
	Name  string      `xml:"id,attr"`
	Type  DataType    `xml:"type,attr"`
	Value interface{} `xml:"value,attr"`
}

func (na NodeAttribute) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	t, _ := na.Type.MarshalXMLAttr(xml.Name{Local: "type"})
	start.Attr = append(start.Attr,
		xml.Attr{
			Name:  xml.Name{Local: "id"},
			Value: na.Name,
		},
		t,
	)
	v, MarshalXML2 := na.Value.(XMLMarshaler)
	v1, MarshalXML := na.Value.(xml.Marshaler)
	if MarshalXML2 {
		v.MarshalXML(e, &start)
	}
	if !(MarshalXML || MarshalXML2) {
		start.Attr = append(start.Attr,
			xml.Attr{
				Name:  xml.Name{Local: "value"},
				Value: na.String(),
			},
		)
	}

	e.EncodeToken(start)

	if MarshalXML {
		e.EncodeElement(v1, xml.StartElement{Name: xml.Name{Local: na.Type.String()}})
	}

	e.EncodeToken(xml.EndElement{
		Name: start.Name,
	})
	return nil
}

func (na NodeAttribute) String() string {
	switch na.Type {
	case DTScratchBuffer:
		// ScratchBuffer is a special case, as its stored as byte[] and ToString() doesn't really do what we want
		if value, ok := na.Value.([]byte); ok {
			return base64.StdEncoding.EncodeToString(value)
		}
		return fmt.Sprint(na.Value)

	case DTDouble:
		v := na.Value.(float64)
		if na.Value == 0 {
			na.Value = 0
		}
		return strconv.FormatFloat(v, 'f', -1, 64)

	case DTFloat:
		v := na.Value.(float32)
		if na.Value == 0 {
			na.Value = 0
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 32)

	default:
		return fmt.Sprint(na.Value)
	}
}

func (na NodeAttribute) GetRows() (int, error) {
	return na.Type.GetRows()
}

func (dt DataType) GetRows() (int, error) {
	switch dt {
	case DTIVec2, DTIVec3, DTIVec4, DTVec2, DTVec3, DTVec4:
		return 1, nil

	case DTMat2:
		return 2, nil

	case DTMat3, DTMat3x4:
		return 3, nil

	case DTMat4x3, DTMat4:
		return 4, nil

	default:
		return 0, errors.New("data type does not have rows")
	}
}

func (na NodeAttribute) GetColumns() (int, error) {
	return na.Type.GetColumns()
}

func (dt DataType) GetColumns() (int, error) {
	switch dt {
	case DTIVec2, DTVec2, DTMat2:
		return 2, nil

	case DTIVec3, DTVec3, DTMat3, DTMat4x3:
		return 3, nil

	case DTIVec4, DTVec4, DTMat3x4, DTMat4:
		return 4, nil

	default:
		return 0, errors.New("data type does not have columns")
	}
}

func (na NodeAttribute) IsNumeric() bool {
	switch na.Type {
	case DTByte, DTShort, DTInt, DTUInt, DTFloat, DTDouble, DTULongLong, DTLong, DTInt8:
		return true
	default:
		return false
	}
}

func (na *NodeAttribute) FromString(str string) error {
	if na.IsNumeric() {
		// Workaround: Some XML files use empty strings, instead of "0" for zero values.
		if str == "" {
			str = "0"
			// Handle hexadecimal integers in XML files
		}
	}

	var (
		err error
	)

	switch na.Type {
	case DTNone:
		// This is a null type, cannot have a value

	case DTByte:
		na.Value = []byte(str)

	case DTShort:

		na.Value, err = strconv.ParseInt(str, 0, 16)
		if err != nil {
			return err
		}

	case DTUShort:
		na.Value, err = strconv.ParseUint(str, 0, 16)
		if err != nil {
			return err
		}

	case DTInt:
		na.Value, err = strconv.ParseInt(str, 0, 32)
		if err != nil {
			return err
		}

	case DTUInt:
		na.Value, err = strconv.ParseUint(str, 0, 16)
		if err != nil {
			return err
		}

	case DTFloat:
		na.Value, err = strconv.ParseFloat(str, 32)
		if err != nil {
			return err
		}

	case DTDouble:
		na.Value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}

	case DTIVec2, DTIVec3, DTIVec4:
		var (
			nums   []string
			length int
		)

		nums = strings.Split(str, ".")
		length, err = na.GetColumns()
		if err != nil {
			return err
		}
		if length != len(nums) {
			return fmt.Errorf("a vector of length %d was expected, got %d", length, len(nums))
		}

		vec := make([]int, length)
		for i, v := range nums {
			var n int64
			n, err = strconv.ParseInt(v, 0, 64)
			vec[i] = int(n)
			if err != nil {
				return err
			}
		}

		na.Value = vec

	case DTVec2, DTVec3, DTVec4:
		var (
			nums   []string
			length int
		)
		nums = strings.Split(str, ".")
		length, err = na.GetColumns()
		if err != nil {
			return err
		}
		if length != len(nums) {
			return fmt.Errorf("a vector of length %d was expected, got %d", length, len(nums))
		}

		vec := make([]float64, length)
		for i, v := range nums {
			vec[i], err = strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
		}

		na.Value = vec

	case DTMat2, DTMat3, DTMat3x4, DTMat4x3, DTMat4:
		// var mat = Matrix.Parse(str);
		// if (mat.cols != na.GetColumns() || mat.rows != na.GetRows()){
		//     return errors.New("Invalid column/row count for matrix");
		// }
		// value = mat;
		return errors.New("not implemented")

	case DTBool:
		na.Value, err = strconv.ParseBool(str)
		if err != nil {
			return err
		}

	case DTString, DTPath, DTFixedString, DTLSString, DTWString, DTLSWString:
		na.Value = str

	case DTTranslatedString:
		// // We'll only set the value part of the translated string, not the TranslatedStringKey / Handle part
		// // That can be changed separately via attribute.Value.Handle
		// if (value == null)
		//     value = new TranslatedString();

		// ((TranslatedString)value).Value = str;

	case DTTranslatedFSString:
		// // We'll only set the value part of the translated string, not the TranslatedStringKey / Handle part
		// // That can be changed separately via attribute.Value.Handle
		// if (value == null)
		//     value = new TranslatedFSString();

		// ((TranslatedFSString)value).Value = str;

	case DTULongLong:
		na.Value, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}

	case DTScratchBuffer:
		na.Value, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}

	case DTLong, DTInt64:
		na.Value, err = strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}

	case DTInt8:
		na.Value, err = strconv.ParseInt(str, 10, 8)
		if err != nil {
			return err
		}

	case DTUUID:
		na.Value, err = uuid.Parse(str)
		if err != nil {
			return err
		}

	default:
		// This should not happen!
		return fmt.Errorf("not implemented for type %v", na.Type)
	}
	return nil
}
