// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

/*
 This package wraps the standard XML library and uses it to build a node tree of
 any document you load. This allows you to look up nodes forwards and backwards,
 as well as perform simple search queries.

 Nodes now simply become collections and don't require you to read them in the
 order in which the xml.Parser finds them.

 The Document currently implements 2 search functions which allow you to
 look for specific nodes.

   *xmlx.Document.SelectNode(namespace, name string) *Node;
   *xmlx.Document.SelectNodes(namespace, name string) []*Node;
   *xmlx.Document.SelectNodesRecursive(namespace, name string) []*Node;

 SelectNode() returns the first, single node it finds matching the given name
 and namespace. SelectNodes() returns a slice containing all the matching nodes
 (without recursing into matching nodes). SelectNodesRecursive() returns a slice
 of all matching nodes, including nodes inside other matching nodes.

 Note that these search functions can be invoked on individual nodes as well.
 This allows you to search only a subset of the entire document.
*/
package xmlx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// This signature represents a character encoding conversion routine.
// Used to tell the xml decoder how to deal with non-utf8 characters.
type CharsetFunc func(charset string, input io.Reader) (io.Reader, error)

// represents a single XML document.
type Document struct {
	Version     string            // XML version
	Encoding    string            // Encoding found in document. If absent, assumes UTF-8.
	StandAlone  string            // Value of XML doctype's 'standalone' attribute.
	Entity      map[string]string // Mapping of custom entity conversions.
	Root        *Node             // The document's root node.
	SaveDocType bool              // Whether not to include the XML doctype in saves.

	useragent string // Used internally
}

// Create a new, empty XML document instance.
func New() *Document {
	return &Document{
		Version:     "1.0",
		Encoding:    "utf-8",
		StandAlone:  "yes",
		SaveDocType: true,
		Entity:      make(map[string]string),
	}
}

// This loads a rather massive table of non-conventional xml escape sequences.
// Needed to make the parser map them to characters properly. It is advised to
// set only those entities needed manually using the document.Entity map, but
// if need be, this method can be called to fill the map with the entire set
// defined on http://www.w3.org/TR/html4/sgml/entities.html
func (this *Document) LoadExtendedEntityMap() { loadNonStandardEntities(this.Entity) }

// Select a single node with the given namespace and name. Returns nil if no
// matching node was found.
func (this *Document) SelectNode(namespace, name string) *Node {
	return this.Root.SelectNode(namespace, name)
}

// Select all nodes with the given namespace and name. Returns an empty slice
// if no matches were found.
// Select all nodes with the given namespace and name, without recursing
// into the children of those matches. Returns an empty slice if no matching
// node was found.
func (this *Document) SelectNodes(namespace, name string) []*Node {
	return this.Root.SelectNodes(namespace, name)
}

// Select all nodes directly under this document, with the given namespace
// and name. Returns an empty slice if no matches were found.
func (this *Document) SelectNodesDirect(namespace, name string) []*Node {
	return this.Root.SelectNodesDirect(namespace, name)
}

// Select all nodes with the given namespace and name, also recursing into the
// children of those matches. Returns an empty slice if no matches were found.
func (this *Document) SelectNodesRecursive(namespace, name string) []*Node {
	return this.Root.SelectNodesRecursive(namespace, name)
}

// Load the contents of this document from the supplied reader.
func (this *Document) LoadStream(r io.Reader, charset CharsetFunc) (err error) {
	xp := xml.NewDecoder(r)
	xp.Entity = this.Entity
	xp.CharsetReader = charset

	this.Root = NewNode(NT_ROOT)
	ct := this.Root

	var tok xml.Token
	var t *Node
	var doctype string

	for {
		if tok, err = xp.Token(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch tt := tok.(type) {
		case xml.SyntaxError:
			return errors.New(tt.Error())
		case xml.CharData:
			t := NewNode(NT_TEXT)
			t.Value = string([]byte(tt))
			ct.AddChild(t)
		case xml.Comment:
			t := NewNode(NT_COMMENT)
			t.Value = strings.TrimSpace(string([]byte(tt)))
			ct.AddChild(t)
		case xml.Directive:
			t = NewNode(NT_DIRECTIVE)
			t.Value = strings.TrimSpace(string([]byte(tt)))
			ct.AddChild(t)
		case xml.StartElement:
			t = NewNode(NT_ELEMENT)
			t.Name = tt.Name
			t.Attributes = make([]*Attr, len(tt.Attr))
			for i, v := range tt.Attr {
				t.Attributes[i] = new(Attr)
				t.Attributes[i].Name = v.Name
				t.Attributes[i].Value = v.Value
			}
			ct.AddChild(t)
			ct = t
		case xml.ProcInst:
			if tt.Target == "xml" { // xml doctype
				doctype = strings.TrimSpace(string(tt.Inst))
				if i := strings.Index(doctype, `standalone="`); i > -1 {
					this.StandAlone = doctype[i+len(`standalone="`) : len(doctype)]
					i = strings.Index(this.StandAlone, `"`)
					this.StandAlone = this.StandAlone[0:i]
				}
			} else {
				t = NewNode(NT_PROCINST)
				t.Target = strings.TrimSpace(tt.Target)
				t.Value = strings.TrimSpace(string(tt.Inst))
				ct.AddChild(t)
			}
		case xml.EndElement:
			if ct = ct.Parent; ct == nil {
				return
			}
		}
	}

	return
}

// Load the contents of this document from the supplied byte slice.
func (this *Document) LoadBytes(d []byte, charset CharsetFunc) (err error) {
	return this.LoadStream(bytes.NewBuffer(d), charset)
}

// Load the contents of this document from the supplied string.
func (this *Document) LoadString(s string, charset CharsetFunc) (err error) {
	return this.LoadStream(strings.NewReader(s), charset)
}

// Load the contents of this document from the supplied file.
func (this *Document) LoadFile(filename string, charset CharsetFunc) (err error) {
	var fd *os.File
	if fd, err = os.Open(filename); err != nil {
		return
	}

	defer fd.Close()
	return this.LoadStream(fd, charset)
}

// Load the contents of this document from the supplied uri using the specifed
// client.
func (this *Document) LoadUriClient(uri string, client *http.Client, charset CharsetFunc) (err error) {
	var r *http.Response

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	if len(this.useragent) > 1 {
		req.Header.Set("User-Agent", this.useragent)
	}

	if r, err = client.Do(req); err != nil {
		return
	}

	defer r.Body.Close()
	return this.LoadStream(r.Body, charset)
}

// Load the contents of this document from the supplied uri.
// (calls LoadUriClient with http.DefaultClient)
func (this *Document) LoadUri(uri string, charset CharsetFunc) (err error) {
	return this.LoadUriClient(uri, http.DefaultClient, charset)
}

// Save the contents of this document to the supplied file.
func (this *Document) SaveFile(path string) error {
	return ioutil.WriteFile(path, this.SaveBytes(), 0600)
}

// Save the contents of this document as a byte slice.
func (this *Document) SaveBytes() []byte {
	var b bytes.Buffer

	if this.SaveDocType {
		b.WriteString(fmt.Sprintf(`<?xml version="%s" encoding="%s" standalone="%s"?>`,
			this.Version, this.Encoding, this.StandAlone))

		if len(IndentPrefix) > 0 {
			b.WriteByte('\n')
		}
	}

	b.Write(this.Root.Bytes())
	return b.Bytes()
}

// Save the contents of this document as a string.
func (this *Document) SaveString() string { return string(this.SaveBytes()) }

// Alias for Document.SaveString(). This one is invoked by anything looking for
// the standard String() method (eg: fmt.Printf("%s\n", mydoc).
func (this *Document) String() string { return string(this.SaveBytes()) }

// Save the contents of this document to the supplied writer.
func (this *Document) SaveStream(w io.Writer) (err error) {
	_, err = w.Write(this.SaveBytes())
	return
}

// Set a custom user agent when making a new request.
func (this *Document) SetUserAgent(s string) {
	this.useragent = s
}
