package lslib

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type XMLMarshaler interface {
	MarshalXML2(e *xml.Encoder, start *xml.StartElement, Type DataType) error
}

type TranslatedString struct {
	Version uint16
	Value   string
	Handle  string
}

func (ts TranslatedString) MarshalXML2(e *xml.Encoder, start *xml.StartElement, Type DataType) error {
	start.Attr = append(start.Attr,
		xml.Attr{
			Name:  xml.Name{Local: "handle"},
			Value: ts.Handle,
		},
		xml.Attr{
			Name:  xml.Name{Local: "version"},
			Value: strconv.Itoa(int(ts.Version)),
		},
		// xml.Attr{
		// 	Name:  xml.Name{Local: "value"},
		// 	Value: ts.Value,
		// },
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

type Vec []float32

func (v Vec) String() string {
	b := &strings.Builder{}
	for _, x := range v {
		b.WriteString(" ")
		if x == 0 {
			x = 0
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'G', 7, 32))
	}
	return b.String()[1:]
}

func (v Vec) MarshalXML2(e *xml.Encoder, start *xml.StartElement, Type DataType) error {
	switch Type {
	case DT_Mat2, DT_Mat3, DT_Mat3x4, DT_Mat4x3, DT_Mat4:
		b := &strings.Builder{}
		for i, x := range v {
			if x < 0 {
				x = float32(roundFloat(float64(x), 2))
			}
			if x == 0 {
				x = 0
			}
			fmt.Fprintf(b, "% .2f", x)
			ii, err := Type.GetColumns()
			if err != nil {
				return err
			}
			if (i % ii) == ii-1 {
				b.WriteString(" \r\n")
			} else {
				b.WriteString(" ")
			}
		}
		// fmt.Fprintln(os.Stderr, b.String()[1:])
		start.Attr = append(start.Attr,
			xml.Attr{
				Name:  xml.Name{Local: "value"},
				Value: b.String(),
			},
		)

	default:
		start.Attr = append(start.Attr,
			xml.Attr{
				Name:  xml.Name{Local: "value"},
				Value: v.String(),
			},
		)
	}
	return nil
}

type DataType int

const (
	DT_None DataType = iota
	DT_Byte
	DT_Short
	DT_UShort
	DT_Int
	DT_UInt
	DT_Float
	DT_Double
	DT_IVec2
	DT_IVec3
	DT_IVec4
	DT_Vec2
	DT_Vec3
	DT_Vec4
	DT_Mat2
	DT_Mat3
	DT_Mat3x4
	DT_Mat4x3
	DT_Mat4
	DT_Bool
	DT_String
	DT_Path
	DT_FixedString
	DT_LSString
	DT_ULongLong
	DT_ScratchBuffer
	// Seems to be unused?
	DT_Long
	DT_Int8
	DT_TranslatedString
	DT_WString
	DT_LSWString
	DT_UUID
	DT_Int64
	DT_TranslatedFSString
	// Last supported datatype, always keep this one at the end
	DT_Max = iota - 1
)

func (dt *DataType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Value: dt.String(),
		Name:  name,
	}, nil
}

func (dt DataType) String() string {
	switch dt {
	case DT_None:
		return "None"
	case DT_Byte:
		return "uint8"
	case DT_Short:
		return "int16"
	case DT_UShort:
		return "uint16"
	case DT_Int:
		return "int32"
	case DT_UInt:
		return "uint32"
	case DT_Float:
		return "float"
	case DT_Double:
		return "double"
	case DT_IVec2:
		return "ivec2"
	case DT_IVec3:
		return "ivec3"
	case DT_IVec4:
		return "ivec4"
	case DT_Vec2:
		return "fvec2"
	case DT_Vec3:
		return "fvec3"
	case DT_Vec4:
		return "fvec4"
	case DT_Mat2:
		return "mat2x2"
	case DT_Mat3:
		return "mat3x3"
	case DT_Mat3x4:
		return "mat3x4"
	case DT_Mat4x3:
		return "mat4x3"
	case DT_Mat4:
		return "mat4x4"
	case DT_Bool:
		return "bool"
	case DT_String:
		return "string"
	case DT_Path:
		return "path"
	case DT_FixedString:
		return "FixedString"
	case DT_LSString:
		return "LSString"
	case DT_ULongLong:
		return "uint64"
	case DT_ScratchBuffer:
		return "ScratchBuffer"
	case DT_Long:
		return "old_int64"
	case DT_Int8:
		return "int8"
	case DT_TranslatedString:
		return "TranslatedString"
	case DT_WString:
		return "WString"
	case DT_LSWString:
		return "LSWString"
	case DT_UUID:
		return "guid"
	case DT_Int64:
		return "int64"
	case DT_TranslatedFSString:
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
	if v, ok := na.Value.(XMLMarshaler); ok {
		v.MarshalXML2(e, &start, na.Type)
	} else {
		start.Attr = append(start.Attr,
			xml.Attr{
				Name:  xml.Name{Local: "value"},
				Value: na.String(),
			},
		)
	}

	e.EncodeToken(start)
	e.EncodeToken(xml.EndElement{
		Name: start.Name,
	})
	return nil
}

func (na NodeAttribute) String() string {
	switch na.Type {
	case DT_ScratchBuffer:
		// ScratchBuffer is a special case, as its stored as byte[] and ToString() doesn't really do what we want
		if value, ok := na.Value.([]byte); ok {
			return base64.StdEncoding.EncodeToString(value)
		}
		return fmt.Sprint(na.Value)

	case DT_Double:
		if na.Value == 0 {
			na.Value = 0
		}
		return fmt.Sprintf("%.15G", na.Value)

	case DT_Float:
		if na.Value == 0 {
			na.Value = 0
		}
		return fmt.Sprintf("%.7G", na.Value)

	default:
		return fmt.Sprint(na.Value)
	}
}

func (na NodeAttribute) GetRows() (int, error) {
	return na.Type.GetRows()
}

func (dt DataType) GetRows() (int, error) {
	switch dt {
	case DT_IVec2, DT_IVec3, DT_IVec4, DT_Vec2, DT_Vec3, DT_Vec4:
		return 1, nil

	case DT_Mat2:
		return 2, nil

	case DT_Mat3, DT_Mat3x4:
		return 3, nil

	case DT_Mat4x3, DT_Mat4:
		return 4, nil

	default:
		return 0, errors.New("Data type does not have rows")
	}
}

func (na NodeAttribute) GetColumns() (int, error) {
	return na.Type.GetColumns()
}

func (dt DataType) GetColumns() (int, error) {
	switch dt {
	case DT_IVec2, DT_Vec2, DT_Mat2:
		return 2, nil

	case DT_IVec3, DT_Vec3, DT_Mat3, DT_Mat4x3:
		return 3, nil

	case DT_IVec4, DT_Vec4, DT_Mat3x4, DT_Mat4:
		return 4, nil

	default:
		return 0, errors.New("Data type does not have columns")
	}
}

func (na NodeAttribute) IsNumeric() bool {
	switch na.Type {
	case DT_Byte, DT_Short, DT_Int, DT_UInt, DT_Float, DT_Double, DT_ULongLong, DT_Long, DT_Int8:
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
	case DT_None:
		// This is a null type, cannot have a value

	case DT_Byte:
		na.Value = []byte(str)

	case DT_Short:

		na.Value, err = strconv.ParseInt(str, 0, 16)
		if err != nil {
			return err
		}

	case DT_UShort:
		na.Value, err = strconv.ParseUint(str, 0, 16)
		if err != nil {
			return err
		}

	case DT_Int:
		na.Value, err = strconv.ParseInt(str, 0, 32)
		if err != nil {
			return err
		}

	case DT_UInt:
		na.Value, err = strconv.ParseUint(str, 0, 16)
		if err != nil {
			return err
		}

	case DT_Float:
		na.Value, err = strconv.ParseFloat(str, 32)
		if err != nil {
			return err
		}

	case DT_Double:
		na.Value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}

	case DT_IVec2, DT_IVec3, DT_IVec4:

		nums := strings.Split(str, ".")
		length, err := na.GetColumns()
		if err != nil {
			return err
		}
		if length != len(nums) {
			return fmt.Errorf("A vector of length %d was expected, got %d", length, len(nums))
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

	case DT_Vec2, DT_Vec3, DT_Vec4:
		nums := strings.Split(str, ".")
		length, err := na.GetColumns()
		if err != nil {
			return err
		}
		if length != len(nums) {
			return fmt.Errorf("A vector of length %d was expected, got %d", length, len(nums))
		}

		vec := make([]float64, length)
		for i, v := range nums {
			vec[i], err = strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
		}

		na.Value = vec

	case DT_Mat2, DT_Mat3, DT_Mat3x4, DT_Mat4x3, DT_Mat4:
		// var mat = Matrix.Parse(str);
		// if (mat.cols != na.GetColumns() || mat.rows != na.GetRows()){
		//     return errors.New("Invalid column/row count for matrix");
		// }
		// value = mat;
		return errors.New("not implemented")

	case DT_Bool:
		na.Value, err = strconv.ParseBool(str)
		if err != nil {
			return err
		}

	case DT_String, DT_Path, DT_FixedString, DT_LSString, DT_WString, DT_LSWString:
		na.Value = str

	case DT_TranslatedString:
		// // We'll only set the value part of the translated string, not the TranslatedStringKey / Handle part
		// // That can be changed separately via attribute.Value.Handle
		// if (value == null)
		//     value = new TranslatedString();

		// ((TranslatedString)value).Value = str;

	case DT_TranslatedFSString:
		// // We'll only set the value part of the translated string, not the TranslatedStringKey / Handle part
		// // That can be changed separately via attribute.Value.Handle
		// if (value == null)
		//     value = new TranslatedFSString();

		// ((TranslatedFSString)value).Value = str;

	case DT_ULongLong:
		na.Value, err = strconv.ParseUint(str, 10, 64)

	case DT_ScratchBuffer:
		na.Value, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}

	case DT_Long, DT_Int64:
		na.Value, err = strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}

	case DT_Int8:
		na.Value, err = strconv.ParseInt(str, 10, 8)
		if err != nil {
			return err
		}

	case DT_UUID:
		na.Value, err = uuid.Parse(str)
		if err != nil {
			return err
		}

	default:
		// This should not happen!
		return fmt.Errorf("FromString() not implemented for type %v", na.Type)
	}
	return nil
}
