package gog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Change struct {
	Title   string
	Changes []string
	Sub     []Change
	recurse bool
}

func (c Change) str(indent int) string {
	var (
		s   = new(strings.Builder)
		ind = strings.Repeat("\t", indent)
	)
	fmt.Fprintln(s, ind+c.Title)
	for _, v := range c.Changes {
		fmt.Fprintln(s, ind+"\t"+v)
	}
	for _, v := range c.Sub {
		s.WriteString(v.str(indent + 1))
	}
	// s.WriteRune('\n')
	return s.String()
}

func (c Change) String() string {
	return c.str(0)
}

func debug(f ...interface{}) {
	if len(os.Args) > 2 && os.Args[2] == "debug" {
		fmt.Println(f...)
	}
}

// Stringify returns the text from a node and all its children
func stringify(h *html.Node) string {
	var (
		f   func(*html.Node)
		str string
		def = h
	)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			str += strings.TrimSpace(n.Data) + " "
		}
		if n.DataAtom == atom.Br {
			str += "\n"
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(def)
	return str
}

// ParseChange returns the parsed change and the next node to start parsing the next change
func ParseChange(h *html.Node, title string) (Change, *html.Node) {
	var c Change
	c.Title = title
	debug("change", strings.TrimSpace(h.Data))
l:
	for h.Data != "hr" {
		switch h.Type {
		case html.ErrorNode:
			panic("An error happened!?!?")
		case html.TextNode:
			debug("text", strings.TrimSpace(h.Data))
			if strings.TrimSpace(h.Data) != "" {
				c.Changes = append(c.Changes, strings.TrimSpace(h.Data))
			}
		case html.ElementNode:
			switch h.DataAtom {
			case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.P:
				debug(h.DataAtom.String())
				if c.Title == "" {
					c.Title = strings.TrimSpace(stringify(h))
				} else {
					tmp := drain(h.NextSibling)
					switch tmp.DataAtom {
					case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
						h = tmp

						c.Changes = append(c.Changes, strings.TrimSpace(stringify(h)))
						break
					}
					var cc Change
					debug("h", strings.TrimSpace(h.Data))
					debug("h", strings.TrimSpace(h.NextSibling.Data))
					cc, h = ParseChange(h.NextSibling, stringify(h))
					debug("h2", h.Data, h.PrevSibling.Data)
					// h = h.Parent
					c.Sub = append(c.Sub, cc)
					continue
				}
			case atom.Ul:
				debug(h.DataAtom.String())
				h = h.FirstChild
				continue
			case atom.Li:
				debug("li", h.DataAtom.String())
				if h.FirstChild != h.LastChild && h.FirstChild.NextSibling.DataAtom != atom.P && h.FirstChild.NextSibling.DataAtom != atom.A {
					var cc Change
					cc, h = ParseChange(h.FirstChild.NextSibling, strings.TrimSpace(h.FirstChild.Data))
					h = h.Parent
					if h.NextSibling != nil {
						h = h.NextSibling
					}
					// pretty.Println(cc)
					c.Sub = append(c.Sub, cc)
					break l
				}
				fallthrough
			default:
				var (
					f   func(*html.Node)
					str string
					def = h
				)
				f = func(n *html.Node) {
					if n.Type == html.TextNode {
						str += strings.TrimSpace(n.Data) + " "
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(def)
				c.Changes = append(c.Changes, strings.TrimSpace(str))
			}
		}
		if h.DataAtom == atom.Ul {
			h = h.Parent
		}
		if h.NextSibling != nil { // Move to the next node, h should never be nil
			debug("next", h.Type, h.NextSibling.Type)
			h = h.NextSibling
		} else if h.DataAtom == atom.Li || h.PrevSibling.DataAtom == atom.Li && h.Parent.DataAtom != atom.Body { // If we are currently in a list then go up one unless that would take us to body
			debug("ul", strings.TrimSpace(h.Data), strings.TrimSpace(h.Parent.Data))
			if h.Parent.NextSibling == nil {
				h = h.Parent
			} else { // go to parents next sibling so we don't parse the same node again
				h = h.Parent.NextSibling
			}
			debug("break2", strings.TrimSpace(h.Data))
			break
		} else {
			debug("I don't believe this should ever happen")
			break
		}
	}
	h = drain(h)
	return c, h
}

// drain skips over non-Element Nodes
func drain(h *html.Node) *html.Node {
	for h.NextSibling != nil {
		if h.Type == html.ElementNode {
			break
		}
		h = h.NextSibling
	}
	return h
}

func ParseChangelog(ch, title string) (Change, error) {
	var (
		p   *html.Node
		v   Change
		err error
	)
	v.Title = title
	p, err = html.Parse(strings.NewReader(ch))
	if err != nil {
		return v, err
	}
	p = p.FirstChild.FirstChild.NextSibling.FirstChild
	for p != nil && p.NextSibling != nil {
		var tmp Change
		tmp, p = ParseChange(p, "")
		if p.DataAtom == atom.Hr {
			p = p.NextSibling
		}
		v.Sub = append(v.Sub, tmp)
	}
	return v, nil
}

func getGOGInfo(id string) (GOGalaxy, error) {
	var (
		r    *http.Response
		err  error
		b    []byte
		info GOGalaxy
	)
	r, err = http.Get("https://api.gog.com/products/" + id + "?expand=downloads,expanded_dlcs,description,screenshots,videos,related_products,changelog")
	if err != nil {
		return GOGalaxy{}, fmt.Errorf("failed to retrieve GOG info: %w", err)
	}
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return GOGalaxy{}, fmt.Errorf("failed to retrieve GOG info: %w", err)
	}
	r.Body.Close()
	err = json.Unmarshal(b, &info)
	if err != nil {
		return GOGalaxy{}, fmt.Errorf("failed to retrieve GOG info: %w", err)
	}
	return info, nil
}
