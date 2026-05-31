// Package save parses and serializes Stardew Valley XML save files.
// Save files are a single-line XML document; we work with a generic node tree
// so that every field is preserved even if not explicitly modeled.
package save

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Node is a generic XML element that preserves attributes, children, and text.
type Node struct {
	Name     string
	Attrs    []xml.Attr
	Children []*Node
	Text     string // non-empty when this is a leaf with text content
}

// Attr returns the value of a named attribute, or "" if absent.
func (n *Node) Attr(local string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// Child returns the first child with the given element name, or nil.
func (n *Node) Child(name string) *Node {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// Children returns all children with the given name.
func (n *Node) ChildrenNamed(name string) []*Node {
	var out []*Node
	for _, c := range n.Children {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// Get traverses a slash-separated path of element names and returns the node,
// or nil if any step is missing. Example: "player/friendshipData"
func (n *Node) Get(path string) *Node {
	parts := strings.SplitN(path, "/", 2)
	if parts[0] == "" || parts[0] == n.Name {
		if len(parts) == 1 || parts[1] == "" {
			return n
		}
		return n.getDown(parts[1])
	}
	return n.getDown(path)
}

func (n *Node) getDown(path string) *Node {
	parts := strings.SplitN(path, "/", 2)
	child := n.Child(parts[0])
	if child == nil {
		return nil
	}
	if len(parts) == 1 {
		return child
	}
	return child.getDown(parts[1])
}

// SetText sets the text of the node at path, creating intermediate nodes if
// they don't exist. Returns an error if the path resolves to a non-leaf node
// that already has children.
func (n *Node) SetText(path, value string) error {
	target := n.Get(path)
	if target == nil {
		return fmt.Errorf("path not found: %s", path)
	}
	if len(target.Children) > 0 {
		return fmt.Errorf("node at %s has children; use a more specific path", path)
	}
	target.Text = value
	return nil
}

// Parse reads an XML document into a Node tree. The root node is returned.
func Parse(r io.Reader) (*Node, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false

	// skip the XML declaration
	var root *Node
	stack := make([]*Node, 0, 32)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml decode: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			node := &Node{
				Name:  t.Name.Local,
				Attrs: append([]xml.Attr(nil), t.Attr...),
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else {
				root = node
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				cur := stack[len(stack)-1]
				cur.Text += text
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("empty document")
	}
	return root, nil
}

// Serialize writes a Node tree back to XML with the standard save file header.
func Serialize(n *Node, w io.Writer) error {
	if _, err := io.WriteString(w, `<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		return err
	}
	return writeNode(n, w)
}

// nsURIToPrefix maps the full namespace URIs used by SDV saves back to the
// short prefixes used in the XML source.  go's xml.Decoder expands prefixes
// to URIs when it stores Attr.Name.Space; we reverse this on output.
var nsURIToPrefix = map[string]string{
	"http://www.w3.org/2001/XMLSchema-instance": "xsi",
	"http://www.w3.org/2001/XMLSchema":          "xsd",
}

func writeNode(n *Node, w io.Writer) error {
	buf := &bytes.Buffer{}

	buf.WriteByte('<')
	buf.WriteString(n.Name)
	for _, a := range n.Attrs {
		buf.WriteByte(' ')
		space := a.Name.Space
		if p, ok := nsURIToPrefix[space]; ok {
			space = p
		}
		if space != "" {
			buf.WriteString(space)
			buf.WriteByte(':')
		}
		buf.WriteString(a.Name.Local)
		buf.WriteString(`="`)
		xml.EscapeText(buf, []byte(a.Value))
		buf.WriteByte('"')
	}

	if len(n.Children) == 0 && n.Text == "" {
		buf.WriteString("/>")
		_, err := w.Write(buf.Bytes())
		return err
	}

	buf.WriteByte('>')
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	if n.Text != "" {
		if err := xml.EscapeText(w, []byte(n.Text)); err != nil {
			return err
		}
	}

	for _, child := range n.Children {
		if err := writeNode(child, w); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "</%s>", n.Name)
	return err
}
