// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xmlx

import (
	"encoding/xml"
	"testing"
)

func TestLoadLocal(t *testing.T) {
	doc := New()

	if err := doc.LoadFile("test.xml", nil); err != nil {
		t.Error(err.Error())
		return
	}

	if len(doc.Root.Children) == 0 {
		t.Errorf("Root node has no children.")
		return
	}
}

func TestWildcard(t *testing.T) {
	doc := New()

	if err := doc.LoadFile("test2.xml", nil); err != nil {
		t.Error(err.Error())
		return
	}

	list := doc.SelectNodes("ns", "*")

	if len(list) != 1 {
		t.Errorf("Wrong number of child elements. Expected 1, got %d.", len(list))
		return
	}
}

func TestWildcardRecursive(t *testing.T) {
	doc := New()

	if err := doc.LoadFile("test2.xml", nil); err != nil {
		t.Error(err.Error())
		return
	}

	list := doc.SelectNodesRecursive("ns", "*")

	if len(list) != 7 {
		t.Errorf("Wrong number of child elements. Expected 7, got %d.", len(list))
		return
	}
}

func _TestLoadRemote(t *testing.T) {
	doc := New()

	if err := doc.LoadUri("http://blog.golang.org/feeds/posts/default", nil); err != nil {
		t.Error(err.Error())
		return
	}

	if len(doc.Root.Children) == 0 {
		t.Errorf("Root node has no children.")
		return
	}
}

func TestSave(t *testing.T) {
	doc := New()

	if err := doc.LoadFile("test.xml", nil); err != nil {
		t.Errorf("LoadFile(): %s", err)
		return
	}

	IndentPrefix = "\t"
	if err := doc.SaveFile("test1.xml"); err != nil {
		t.Errorf("SaveFile(): %s", err)
		return
	}
}

func TestNodeSearch(t *testing.T) {
	doc := New()

	if err := doc.LoadFile("test1.xml", nil); err != nil {
		t.Errorf("LoadFile(): %s", err)
		return
	}

	if node := doc.SelectNode("", "item"); node == nil {
		t.Errorf("SelectNode(): No node found.")
		return
	}

	nodes := doc.SelectNodes("", "item")
	if len(nodes) == 0 {
		t.Errorf("SelectNodes(): no nodes found.")
		return
	}

	ch := doc.SelectNode("", "channel")
	// Test that SelectNodes properly selects multiple nodes
	links := ch.SelectNodes("", "link")
	if len(links) != 8 {
		t.Errorf("SelectNodes(): Expected 8, Got %d", len(links))
		return
	}
}

type Image struct {
	Title       string `xml:"title"`
	Url         string `xml:"url"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Width       int    `xml:"width"`
	Height      int    `xml:"height"`
}

func TestUnmarshal(t *testing.T) {
	doc := New()
	err := doc.LoadFile("test1.xml", nil)

	if err != nil {
		t.Errorf("LoadFile(): %s", err)
		return
	}

	node := doc.SelectNode("", "image")
	if node == nil {
		t.Errorf("SelectNode(): No node found.")
		return
	}

	var img Image
	if err = node.Unmarshal(&img); err != nil {
		t.Errorf("Unmarshal(): %s", err)
		return
	}

	if img.Title != "WriteTheWeb" {
		t.Errorf("Image.Title has incorrect value. Got '%s', expected 'WriteTheWeb'.", img.Title)
		return
	}
}

func TestStringNamespaces(t *testing.T) {
	doc := New()
	err := doc.LoadFile("test3.xml", nil)

	if err != nil {
		t.Errorf("LoadFile(): %s", err)
		return
	}

	expected := `<root xmlns:foo="http:/example.org/foo">
  <child foo:bar="1">
    <grandchild xmlns:foo="">
      <great-grandchild bar="2">&#xA;      </great-grandchild>
    </grandchild>
  </child>
</root>
`

	if got := doc.Root.String(); got != expected {
		t.Fatalf("expected: %s\ngot: %s\n", expected, got)
	}
}

func TestStringEscaping(t *testing.T) {
	doc := New()
	err := doc.LoadFile("test4.xml", nil)

	if err != nil {
		t.Errorf("LoadFile(): %s", err)
		return
	}

	expected := `<body>  &lt;https://example.com/file/fm/SU0vRk0xLzIwMTMwOTEwLzA1MDA0MS5ybXdhdGVzdEByZXV0ZXJzLmNvbTEzNzg4NDU1OTk4OTA/Screen%20Shot%202013-09-10%20at%2021.33.54.png&gt; File Attachment:-Screen Shot 2013-09-10 at 21.33.54.png  </body>
`

	if got := doc.Root.String(); got != expected {
		t.Fatalf("expected: %s\ngot: %s\n", expected, got)
	}
}

func TestElementNodeValueFetch(t *testing.T) {
	data := `<car><color>
	r<cool />
	ed</color><brand>BMW</brand><price>50
	<cheap />.25</price><count>6
	<small />2
	</count><available>
	Tr
	<found />
	ue</available></car>`
	doc := New()

	if err := doc.LoadString(data, nil); nil != err {
		t.Fatalf("LoadString(): %s", err)
	}

	carN := doc.SelectNode("", "car")
	if v := carN.S("", "brand"); v != "BMW" {
		t.Errorf("Failed to get brand as string, got: '%s', wanted: 'BMW'", v)
	}
	if v := carN.S("", "color"); v != "red" {
		t.Errorf("Failed to get color as string, got: '%s', wanted: 'red'", v)
	}

	if v := carN.I("", "count"); v != 62 {
		t.Errorf("Failed to get count using I, got: %v, wanted: 62", v)
	}
	if v := carN.I8("", "count"); v != 62 {
		t.Errorf("Failed to get count using I8, got: %v, wanted: 62", v)
	}
	if v := carN.I16("", "count"); v != 62 {
		t.Errorf("Failed to get count using I16, got: %v, wanted: 62", v)
	}
	if v := carN.I32("", "count"); v != 62 {
		t.Errorf("Failed to get count using I32, got: %v, wanted: 62", v)
	}
	if v := carN.I64("", "count"); v != 62 {
		t.Errorf("Failed to get count using I64, got: %v, wanted: 62", v)
	}
	if v := carN.U("", "count"); v != 62 {
		t.Errorf("Failed to get count using U, got: %v, wanted: 62", v)
	}
	if v := carN.U8("", "count"); v != 62 {
		t.Errorf("Failed to get count using U8, got: %v, wanted: 62", v)
	}
	if v := carN.U16("", "count"); v != 62 {
		t.Errorf("Failed to get count using U16, got: %v, wanted: 62", v)
	}
	if v := carN.U32("", "count"); v != 62 {
		t.Errorf("Failed to get count using U32, got: %v, wanted: 62", v)
	}
	if v := carN.U64("", "count"); v != 62 {
		t.Errorf("Failed to get count using U64, got: %v, wanted: 62", v)
	}

	if v := carN.F32("", "price"); v != 50.25 {
		t.Errorf("Failed to get price using F32, got: %v, wanted: 50.25", v)
	}
	if v := carN.F64("", "price"); v != 50.25 {
		t.Errorf("Failed to get price using F64, got: %v, wanted: 50.25", v)
	}

	if v := carN.B("", "available"); v != true {
		t.Errorf("Failed to get availability using B, got: %v, wanted: true", v)
	}
}

// node.SetValue(x); x == node.GetValue
func TestElementNodeValueFetchAndSetIdentity(t *testing.T) {
	// Setup: <root><text>xyzzy</text></root>
	// The xmlx parser creates a nameless NT_TEXT node containing the value 'xyzzy'
	rootN := NewNode(NT_ROOT)
	rootN.Name = xml.Name{Space: "", Local: "root"}
	textN := NewNode(NT_ELEMENT)
	textN.Name = xml.Name{Space: "", Local: "text"}
	namelessN := NewNode(NT_TEXT)
	namelessN.Value = "xyzzy"
	rootN.AddChild(textN)
	textN.AddChild(namelessN)

	targetN := rootN.SelectNode("", "text") // selects textN
	if targetN != textN {
		t.Errorf("Failed to get the correct textN, got %#v", targetN)
	}

	// targetN.Value is empty (as the value lives in the childNode)
	if targetN.Value != "" {
		t.Errorf("Failed to prepare correctly, TargetN.Value is not empty, it contains %#v", targetN.Value)
	}

	// Test correct retrieval
	if v := rootN.S("", "text"); v != "xyzzy" {
		t.Errorf("Failed to get value as string, got: '%s', wanted: 'xyzzy'", v)
	}

	// Set the value of the nameless child
	targetN.SetValue("plugh")

	// Test correct retrieval
	if v := rootN.S("", "text"); v != "plugh" {
		t.Errorf("Failed to get value as string, got: '%s', wanted: 'plugh'", v)
	}
}

// Test as it could be used to read in a XML file, change some values and write it out again.
// For example, a HTML/XML proxy service.
func TestElementNodeValueFetchAndSet(t *testing.T) {
	IndentPrefix = ""

	data := `<car><color>
	r<cool />
	ed</color><brand>BM<awesome />W</brand><price>50
	<cheap />.25</price><count>6
	<small />2
	</count><available>
	Tr
	<found />
	ue</available></car>`
	doc := New()

	if err := doc.LoadString(data, nil); nil != err {
		t.Fatalf("LoadString(): %s", err)
	}

	carN := doc.SelectNode("", "car")
	if carN == nil {
		t.Fatalf("Failed to get the car, got nil, wanted Node{car}")
	}

	colorNode := carN.SelectNode("", "color")
	if colorNode == nil {
		t.Fatalf("Failed to get the color, got nil, wanted Node{color}")
	}

	colorVal := colorNode.GetValue()
	if colorVal != "red" {
		t.Fatalf("Failed to get the color, got %v, wanted 'red'", colorVal)
	}

	colorNode.SetValue("blue")

	expected := `<car><color>blue</color><brand>BM<awesome />W</brand><price>50
	<cheap />.25</price><count>6
	<small />2
	</count><available>
	Tr
	<found />
	ue</available></car>`

	if got := doc.Root.String(); got != expected {
		t.Fatalf("expected: \n%s\ngot: \n%s\n", expected, got)
	}

	// now set the brand
	brandNode := carN.SelectNode("", "brand")
	if brandNode == nil {
		t.Fatalf("Failed to get the color, got nil, wanted Node{brand}")
	}

	brandVal := brandNode.GetValue()
	if brandVal != "BMW" {
		t.Fatalf("Failed to get the brand, got %v, wanted 'BMW'", brandVal)
	}

	brandNode.SetValue("Trabant")

	// Notice, we lose the <awesome /> tag in BMW, that's intentional
	expected = `<car><color>blue</color><brand>Trabant</brand><price>50
	<cheap />.25</price><count>6
	<small />2
	</count><available>
	Tr
	<found />
	ue</available></car>`

	if got := doc.Root.String(); got != expected {
		t.Fatalf("expected: \n%s\ngot: \n%s\n", expected, got)
	}
}

func TestSelectNodesDirect(t *testing.T) {
	data := `<root><wrapper><hidden></hidden>
	<hidden></hidden></wrapper></root>`
	doc := New()

	if err := doc.LoadString(data, nil); nil != err {
		t.Fatalf("LoadString(): %s", err)
	}

	root := doc.SelectNode("*", "root")
	if root == nil {
		t.Fatalf("Failed to get root, got nil, wanted Node{root}")
	}

	nodes := root.SelectNodesDirect("*", "hidden")

	if len(nodes) != 0 {
		t.Errorf("SelectDirectNodes should not select children of children.")
	}

	wrapper := root.SelectNode("*", "wrapper")
	if wrapper == nil {
		t.Fatalf("Failed to get wrapper, got nil, wanted Node{wrapper}")
	}

	nodes = wrapper.SelectNodesDirect("*", "hidden")
	if len(nodes) != 2 {
		t.Errorf("Unexcepted hidden nodes found. Expected: 2, Got: %d", len(nodes))
	}
}
