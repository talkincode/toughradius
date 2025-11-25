## XMLX

*NOTICE*

This is forked repository that was provided from github.com/jteeuwen/go-pkg-xmlx.

See https://www.reddit.com/r/golang/comments/7vv9zz/popular_lib_gobindata_removed_from_github_or_why/

This package wraps the standard XML library and uses it to build a node tree of
any document you load. This allows you to look up nodes forwards and backwards,
as well as perform search queries (no xpath support).

Nodes now simply become collections and don't require you to read them in the
order in which the xml.Parser finds them.


### Dependencies

None.


### API

The Document currently implements 2 simple search functions which allow you to
look for specific nodes.

    *document.SelectNode(namespace, name string) *Node;
    *document.SelectNodes(namespace, name string) []*Node;
 
`SelectNode()` returns the first, single node it finds matching the given name
and namespace. `SelectNodes()` returns a slice containing all the matching nodes.

Note that these search functions can be invoked on individual nodes as well.
This allows you to search only a subset of the entire document.

Each node exposes also a number of functions which allow easy access to a node
value or an attribute value. They come in various forms to allow transparent
conversion to types: int, int64, uint, uint64, float32, float64:

    *node.S(ns, name string) string
    *node.I(ns, name string) int
    *node.I8(ns, name string) int8
    *node.I16(ns, name string) int16
    *node.I32(ns, name string) int32
    *node.I64(ns, name string) int64
    *node.U(ns, name string) uint
    *node.U8(ns, name string) uint8
    *node.U16(ns, name string) uint16
    *node.U32(ns, name string) uint32
    *node.U64(ns, name string) uint64
    *node.F32(ns, name string) float32
    *node.F64(ns, name string) float64
    *node.B(ns, name string) bool

Note that these functions actually consider child nodes for matching names as
well as the current node. In effect they first perform a node.SelectNode() and
then return the value of the resulting node converted to the appropriate type.
This allows you to do this:

Consider this piece of xml:

    <car>
       <color>red</color>
       <brand>BMW</brand>
    </car>

Now this code:

    node := doc.SelectNode("", "car")
    brand := node.S("", "brand")

Eventhough `brand` is not the name of `node`, we still get the right value
back (BMW), because `node.S()` searches through the child nodes when looking
for the value if the current node does not match the given namespace and
name.

For attributes, we only go through the attributes of the current node this
function is invoked on:

    *node.As(ns, name string) string
    *node.Ai(ns, name string) int
    *node.Ai8(ns, name string) int8
    *node.Ai16(ns, name string) int16
    *node.Ai32(ns, name string) int32
    *node.Ai64(ns, name string) int64
    *node.Au(ns, name string) uint
    *node.Au8(ns, name string) uint8
    *node.Au16(ns, name string) uint16
    *node.Au32(ns, name string) uint32
    *node.Au64(ns, name string) uint64
    *node.Af32(ns, name string) float32
    *node.Af64(ns, name string) float64
    *node.Ab(ns, name string) bool

All of these functions return either "" or 0 when the specified node or
attribute could not be found. No errors are generated.

The namespace name specified in the functions above must either match the
namespace you expect a node/attr to have, or you can specify a wildcard `*`.
This makes node searches easier in case you do not care what namespace name
there is or if there is one at all. Node and attribute names as well, may
be supplied as the wildcard `*`. This allows us to fetch all child nodes for
a given namespace, regardless of their names.

All numeric type-conversion methods assume base-10 numbers data.


### License

This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
license. Its contents can be found in the LICENSE file.

