// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xmlx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const (
	NT_ROOT = iota
	NT_DIRECTIVE
	NT_PROCINST
	NT_COMMENT
	NT_TEXT
	NT_ELEMENT
)

// IndentPrefix holds the value for a single identation level, if one
// chooses to want indentation in the node.String() and node.Bytes() output.
// This would normally be set to a single tab, or a number of spaces.
var IndentPrefix = ""

type Attr struct {
	Name  xml.Name // Attribute namespace and name.
	Value string   // Attribute value.
}

type Node struct {
	Type       byte     // Node type.
	Name       xml.Name // Node namespace and name.
	Children   []*Node  // Child nodes.
	Attributes []*Attr  // Node attributes.
	Parent     *Node    // Parent node.
	Value      string   // Node value.
	Target     string   // procinst field.
}

func NewNode(tid byte) *Node {
	n := new(Node)
	n.Type = tid
	n.Children = make([]*Node, 0, 10)
	n.Attributes = make([]*Attr, 0, 10)
	return n
}

// This wraps the standard xml.Unmarshal function and supplies this particular
// node as the content to be unmarshalled.
func (this *Node) Unmarshal(obj interface{}) error {
	return xml.NewDecoder(bytes.NewBuffer(this.bytes())).Decode(obj)
}

func (this *Node) GetValue() string {
	res := ""
	for _, node := range this.Children {
		if node.Type == NT_TEXT {
			res += strings.TrimSpace(node.Value)
		}
	}

	return res
}

// SetValue sets the value of the node to the given parameter.
// It deletes all children of the node so the old data does not get back at node.GetValue
func (this *Node) SetValue(val string) {
	t := NewNode(NT_TEXT)
	t.Value = val
	t.Parent = this
	this.Children = []*Node{t} // brutally replace all other children
}

// Get node value as string
func (this *Node) S(namespace, name string) string {
	foundNode := rec_SelectNode(this, namespace, name)
	if foundNode == nil {
		return ""
	} else {
		return foundNode.GetValue()
	}
}

// Get node value as int
func (this *Node) I(namespace, name string) int {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 0)
		return int(n)
	}
	return 0
}

// Get node value as int8
func (this *Node) I8(namespace, name string) int8 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 8)
		return int8(n)
	}
	return 0
}

// Get node value as int16
func (this *Node) I16(namespace, name string) int16 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 16)
		return int16(n)
	}
	return 0
}

// Get node value as int32
func (this *Node) I32(namespace, name string) int32 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 32)
		return int32(n)
	}
	return 0
}

// Get node value as int64
func (this *Node) I64(namespace, name string) int64 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 64)
		return n
	}
	return 0
}

// Get node value as uint
func (this *Node) U(namespace, name string) uint {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseUint(value, 10, 0)
		return uint(n)
	}
	return 0
}

// Get node value as uint8
func (this *Node) U8(namespace, name string) uint8 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseUint(value, 10, 8)
		return uint8(n)
	}
	return 0
}

// Get node value as uint16
func (this *Node) U16(namespace, name string) uint16 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseUint(value, 10, 16)
		return uint16(n)
	}
	return 0
}

// Get node value as uint32
func (this *Node) U32(namespace, name string) uint32 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseUint(value, 10, 32)
		return uint32(n)
	}
	return 0
}

// Get node value as uint64
func (this *Node) U64(namespace, name string) uint64 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseUint(value, 10, 64)
		return n
	}
	return 0
}

// Get node value as float32
func (this *Node) F32(namespace, name string) float32 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseFloat(value, 32)
		return float32(n)
	}
	return 0
}

// Get node value as float64
func (this *Node) F64(namespace, name string) float64 {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseFloat(value, 64)
		return n
	}
	return 0
}

// Get node value as bool
func (this *Node) B(namespace, name string) bool {
	value := this.S(namespace, name)
	if value != "" {
		n, _ := strconv.ParseBool(value)
		return n
	}
	return false
}

// Get attribute value as string
func (this *Node) As(namespace, name string) string {
	for _, v := range this.Attributes {
		if (namespace == "*" || namespace == v.Name.Space) && name == v.Name.Local {
			return v.Value
		}
	}
	return ""
}

// Get attribute value as int
func (this *Node) Ai(namespace, name string) int {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseInt(s, 10, 0)
		return int(n)
	}
	return 0
}

// Get attribute value as int8
func (this *Node) Ai8(namespace, name string) int8 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseInt(s, 10, 8)
		return int8(n)
	}
	return 0
}

// Get attribute value as int16
func (this *Node) Ai16(namespace, name string) int16 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseInt(s, 10, 16)
		return int16(n)
	}
	return 0
}

// Get attribute value as int32
func (this *Node) Ai32(namespace, name string) int32 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseInt(s, 10, 32)
		return int32(n)
	}
	return 0
}

// Get attribute value as int64
func (this *Node) Ai64(namespace, name string) int64 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseInt(s, 10, 64)
		return n
	}
	return 0
}

// Get attribute value as uint
func (this *Node) Au(namespace, name string) uint {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseUint(s, 10, 0)
		return uint(n)
	}
	return 0
}

// Get attribute value as uint8
func (this *Node) Au8(namespace, name string) uint8 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseUint(s, 10, 8)
		return uint8(n)
	}
	return 0
}

// Get attribute value as uint16
func (this *Node) Au16(namespace, name string) uint16 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseUint(s, 10, 16)
		return uint16(n)
	}
	return 0
}

// Get attribute value as uint32
func (this *Node) Au32(namespace, name string) uint32 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseUint(s, 10, 32)
		return uint32(n)
	}
	return 0
}

// Get attribute value as uint64
func (this *Node) Au64(namespace, name string) uint64 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseUint(s, 10, 64)
		return n
	}
	return 0
}

// Get attribute value as float32
func (this *Node) Af32(namespace, name string) float32 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseFloat(s, 32)
		return float32(n)
	}
	return 0
}

// Get attribute value as float64
func (this *Node) Af64(namespace, name string) float64 {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseFloat(s, 64)
		return n
	}
	return 0
}

// Get attribute value as bool
func (this *Node) Ab(namespace, name string) bool {
	s := this.As(namespace, name)
	if s != "" {
		n, _ := strconv.ParseBool(s)
		return n
	}
	return false
}

// Returns true if this node has the specified attribute. False otherwise.
func (this *Node) HasAttr(namespace, name string) bool {
	for _, v := range this.Attributes {
		if namespace != "*" && namespace != v.Name.Space {
			continue
		}

		if name == "*" || name == v.Name.Local {
			return true
		}
	}

	return false
}

// Select single node by name
func (this *Node) SelectNode(namespace, name string) *Node {
	return rec_SelectNode(this, namespace, name)
}

func rec_SelectNode(cn *Node, namespace, name string) *Node {
	if (namespace == "*" || cn.Name.Space == namespace) && (name == "*" || cn.Name.Local == name) {
		return cn
	}

	var tn *Node
	for _, v := range cn.Children {
		if tn = rec_SelectNode(v, namespace, name); tn != nil {
			return tn
		}
	}

	return nil
}

// Select multiple nodes by name
func (this *Node) SelectNodes(namespace, name string) []*Node {
	list := make([]*Node, 0, 16)
	rec_SelectNodes(this, namespace, name, &list, false)
	return list
}

// Select multiple nodes directly under this node, by name.
func (this *Node) SelectNodesDirect(namespace, name string) []*Node {
	list := make([]*Node, 0, 16)

	for _, v := range this.Children {
		if (namespace == "*" || v.Name.Space == namespace) && (name == "*" || v.Name.Local == name) {
			list = append(list, v)
		}
	}

	return list
}

// Select multiple nodes by name
func (this *Node) SelectNodesRecursive(namespace, name string) []*Node {
	list := make([]*Node, 0, 16)
	rec_SelectNodes(this, namespace, name, &list, true)
	return list
}

func rec_SelectNodes(cn *Node, namespace, name string, list *[]*Node, recurse bool) {
	if (namespace == "*" || cn.Name.Space == namespace) && (name == "*" || cn.Name.Local == name) {
		*list = append(*list, cn)
		if !recurse {
			return
		}
	}

	for _, v := range cn.Children {
		rec_SelectNodes(v, namespace, name, list, recurse)
	}
}

func (this *Node) RemoveNameSpace() {
	this.Name.Space = ""
	//	this.RemoveAttr("xmlns") //This is questionable

	for _, v := range this.Children {
		v.RemoveNameSpace()
	}
}

func (this *Node) RemoveAttr(name string) {
	for i, v := range this.Attributes {
		if name == v.Name.Local {
			//Delete it
			this.Attributes = append(this.Attributes[:i], this.Attributes[i+1:]...)
		}
	}
}

func (this *Node) SetAttr(name, value string) {
	for _, v := range this.Attributes {
		if name == v.Name.Local {
			v.Value = value
			return
		}
	}
	//Add
	attr := new(Attr)
	attr.Name.Local = name
	attr.Name.Space = ""
	attr.Value = value
	this.Attributes = append(this.Attributes, attr)
	return
}

// Convert node to appropriate []byte representation based on it's @Type.
// Note that NT_ROOT is a special-case empty node used as the root for a
// Document. This one has no representation by itself. It merely forwards the
// String() call to it's child nodes.
func (this *Node) Bytes() []byte { return this.bytes() }

func (this *Node) bytes() (b []byte) {
	switch this.Type {
	case NT_PROCINST:
		b = this.printProcInst()
	case NT_COMMENT:
		b = this.printComment()
	case NT_DIRECTIVE:
		b = this.printDirective()
	case NT_ELEMENT:
		b = this.printElement()
	case NT_TEXT:
		b = this.printText()
	case NT_ROOT:
		b = this.printRoot()
	}
	return
}

// Convert node to appropriate string representation based on it's @Type.
// Note that NT_ROOT is a special-case empty node used as the root for a
// Document. This one has no representation by itself. It merely forwards the
// String() call to it's child nodes.
func (this *Node) String() (s string) {
	return string(this.bytes())
}

func (this *Node) printRoot() []byte {
	var b bytes.Buffer
	for _, v := range this.Children {
		b.Write(v.bytes())
	}
	return b.Bytes()
}

func (this *Node) printProcInst() []byte {
	return []byte("<?" + this.Target + " " + this.Value + "?>")
}

func (this *Node) printComment() []byte {
	return []byte("<!-- " + this.Value + " -->")
}

func (this *Node) printDirective() []byte {
	return []byte("<!" + this.Value + "!>")
}

func (this *Node) printText() []byte {
	val := []byte(this.Value)
	if len(this.Parent.Children) > 1 {
		return val
	}
	var b bytes.Buffer
	xml.EscapeText(&b, val)
	return b.Bytes()
}

func (this *Node) printElement() []byte {
	var b bytes.Buffer

	if len(this.Name.Space) > 0 {
		b.WriteRune('<')
		b.WriteString(this.Name.Space)
		b.WriteRune(':')
		b.WriteString(this.Name.Local)
	} else {
		b.WriteRune('<')
		b.WriteString(this.Name.Local)
	}

	for _, v := range this.Attributes {
		if len(v.Name.Space) > 0 {
			prefix := this.spacePrefix(v.Name.Space)
			b.WriteString(fmt.Sprintf(` %s:%s="%s"`, prefix, v.Name.Local, v.Value))
		} else {
			b.WriteString(fmt.Sprintf(` %s="%s"`, v.Name.Local, v.Value))
		}
	}

	if len(this.Children) == 0 && len(this.Value) == 0 {
		b.WriteString(" />")
		return b.Bytes()
	}

	b.WriteRune('>')

	for _, v := range this.Children {
		b.Write(v.bytes())
	}

	xml.EscapeText(&b, []byte(this.Value))
	if len(this.Name.Space) > 0 {
		b.WriteString("</")
		b.WriteString(this.Name.Space)
		b.WriteRune(':')
		b.WriteString(this.Name.Local)
		b.WriteRune('>')
	} else {
		b.WriteString("</")
		b.WriteString(this.Name.Local)
		b.WriteRune('>')
	}

	return b.Bytes()
}

// spacePrefix resolves the given space (e.g. a url) to the prefix it was
// assigned by an attribute by the current node, or one of its parents.
func (this *Node) spacePrefix(space string) string {
	for _, attr := range this.Attributes {
		if attr.Name.Space == "xmlns" && attr.Value == space {
			return attr.Name.Local
		}
	}
	if this.Parent == nil {
		return space
	}
	return this.Parent.spacePrefix(space)
}

// Add a child node
func (this *Node) AddChild(t *Node) {
	if t.Parent != nil {
		t.Parent.RemoveChild(t)
	}
	t.Parent = this
	this.Children = append(this.Children, t)
}

// Remove a child node
func (this *Node) RemoveChild(t *Node) {
	p := -1
	for i, v := range this.Children {
		if v == t {
			p = i
			break
		}
	}

	if p == -1 {
		return
	}

	copy(this.Children[p:], this.Children[p+1:])
	this.Children = this.Children[0 : len(this.Children)-1]

	t.Parent = nil
}
