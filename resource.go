package lslib

import (
	"encoding/xml"
	"io"
)

type LSMetadata struct {
	//public const uint CurrentMajorVersion = 33;

	Timestamp uint64 `xml:"-"`
	Major     uint32 `xml:"major,attr"`
	Minor     uint32 `xml:"minor,attr"`
	Revision  uint32 `xml:"revision,attr"`
	Build     uint32 `xml:"build,attr"`
}

type format struct {
	name, magic string
	read        func(io.Reader) (Resource, error)
}

type Resource struct {
	Metadata LSMetadata `xml:"version"`
	Regions  []*Node    `xml:"region"`
}

func (r *Resource) Read(io.Reader) {

}

// public Resource()
// {
//     Metadata.MajorVersion = 3;
// }

type Node struct {
	Name       string          `xml:"id,attr"`
	Parent     *Node           `xml:"-"`
	Attributes []NodeAttribute `xml:"attribute"`
	Children   []*Node         `xml:"children>node,omitempty"`

	RegionName string `xml:"-"`
}

func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	R := xml.Name{
		Local: "region",
	}
	N := xml.Name{
		Local: "node",
	}
	I := xml.Name{
		Local: "id",
	}
	C := xml.Name{
		Local: "children",
	}
	if n.RegionName != "" {
		tmp := xml.StartElement{
			Name: R,
			Attr: []xml.Attr{{Name: I, Value: n.RegionName}},
		}
		e.EncodeToken(tmp)
	}
	e.EncodeToken(xml.StartElement{
		Name: N,
		Attr: []xml.Attr{{Name: I, Value: n.Name}},
	})
	e.EncodeElement(n.Attributes, xml.StartElement{Name: xml.Name{Local: "attribute"}})
	if len(n.Children) > 0 {
		e.EncodeToken(xml.StartElement{Name: C})
		e.Encode(n.Children)
		e.EncodeToken(xml.EndElement{Name: C})
	}
	e.EncodeToken(xml.EndElement{Name: N})
	if n.RegionName != "" {
		e.EncodeToken(xml.EndElement{Name: R})
	}
	return nil
}

func (n Node) ChildCount() (sum int) {
	// for _, v := range n.Children {
	// 	sum += len(v)
	// }
	return len(n.Children)
}

func (n *Node) AppendChild(child *Node) {
	n.Children = append(n.Children, child)
}

//      int TotalChildCount()
// {
//     int count = 0;
//     foreach (var key in Children)
//     {
//         foreach (var child in key.Value)
//         {
//             count += 1 + child.TotalChildCount();
//         }
//     }

//     return count;
// }
