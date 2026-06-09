package handlers

import (
	"net"

	"github.com/talkincode/toughradius/v9/pkg/common"
)

// ipv6PrefixOrNA renders an IPv6 prefix attribute (RFC 3162 Framed-IPv6-Prefix,
// RFC 4818 Delegated-IPv6-Prefix) as a CIDR string, returning common.NA when the
// attribute is absent. A nil *net.IPNet stringifies to "<nil>", which is
// meaningless to persist for later filtering and display, so it is normalized to
// the project's not-available sentinel.
func ipv6PrefixOrNA(prefix *net.IPNet) string {
	if prefix == nil {
		return common.NA
	}
	return prefix.String()
}

// ipv6AddrOrNA renders an IPv6 host-address attribute (RFC 6911
// Framed-IPv6-Address) as a string, returning common.NA when the attribute is
// absent or the unspecified address. A nil net.IP stringifies to "<nil>", which
// is normalized to the project's not-available sentinel.
func ipv6AddrOrNA(ip net.IP) string {
	if len(ip) == 0 || ip.IsUnspecified() {
		return common.NA
	}
	return ip.String()
}
