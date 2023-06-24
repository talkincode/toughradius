package cwmp

import (
	"github.com/talkincode/toughradius/v8/common/xmlx"
)

func getDocNodeValue(doc *xmlx.Document, ns string, name string) string {
	node := doc.SelectNode(ns, name)
	if node != nil {
		return node.GetValue()
	}
	return ""
}

func getNodeValue(node *xmlx.Node, ns string, name string) string {
	_node := node.SelectNode(ns, name)
	if _node != nil {
		return _node.GetValue()
	}
	return ""
}
