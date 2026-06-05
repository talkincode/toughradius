package vendorparsers

import (
	"regexp"
	"strconv"
)

var (
	// vlanStdRegexp1 matches the "slot/subslot/port:vlan[.vlan2]" NAS-Port-Id
	// encoding, e.g. "3/0/1:2814.727".
	vlanStdRegexp1 = regexp.MustCompile(`\w?\s?\d+/\d+/\d+:(\d+)(\.(\d+))?\s?`)
	// vlanStdRegexp2 matches the "vlanid=<n>;vlanid2=<n>;" NAS-Port-Id encoding,
	// e.g. "slot=2;subslot=2;port=22;vlanid=503;vlanid2=100;".
	vlanStdRegexp2 = regexp.MustCompile(`vlanid=(\d+);(vlanid2=?(\d+);)?`)
)

// ParseVlanIDs extracts the inner and outer VLAN IDs encoded in a NAS-Port-Id
// string. It recognizes the two common encodings:
//
//   - "slot/subslot/port:vlan[.vlan2]"  (e.g. "3/0/1:2814.727")
//   - "vlanid=<n>;vlanid2=<n>;"         (e.g. "vlanid=503;vlanid2=100;")
//
// It returns (0, 0) when the string carries no recognizable VLAN information.
// This shared helper lets every vendor parser extract VLANs consistently
// instead of re-implementing (or stubbing out) the logic.
func ParseVlanIDs(nasPortID string) (vlanid1, vlanid2 int64) {
	attrs := vlanStdRegexp1.FindStringSubmatch(nasPortID)
	if attrs == nil {
		attrs = vlanStdRegexp2.FindStringSubmatch(nasPortID)
	}
	if attrs != nil {
		vlanid1, _ = strconv.ParseInt(attrs[1], 10, 64)
		if attrs[2] != "" {
			vlanid2, _ = strconv.ParseInt(attrs[3], 10, 64)
		}
	}
	return vlanid1, vlanid2
}
