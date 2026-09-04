package waf

import (
	"bytes"
	"fmt"
	"net"
	"sort"
)

type Ranger interface {
	Lookup(net.IP) (r Range, ok bool)
	NumRanges() int
}

type IPList struct {
	ranges []Range
}

type Range struct {
	First, Last net.IP
	Description string
}

func (r Range) String() string {
	return fmt.Sprintf("%s-%s: %s", r.First, r.Last, r.Description)
}

// New creates an IP list. Ranges are sorted by lower-bound IP.
// Overlapping ranges: first matching range in sorted order wins.
func New(ranges []Range) *IPList {
	sorted := append([]Range(nil), ranges...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i].First, sorted[j].First
		c := bytes.Compare(a, b)
		if c != 0 {
			return c < 0
		}
		return bytes.Compare(sorted[i].Last, sorted[j].Last) < 0
	})
	return &IPList{ranges: sorted}
}

func (ipl *IPList) NumRanges() int {
	if ipl == nil {
		return 0
	}
	return len(ipl.ranges)
}

// Lookup returns the range the given IP is in. ok is false if no range is found
// or if ip is nil / unparseable.
func (ipl *IPList) Lookup(ip net.IP) (r Range, ok bool) {
	if ipl == nil || ip == nil {
		return
	}
	v4 := ip.To4()
	if v4 != nil {
		return ipl.lookup(v4)
	}
	v6 := ip.To16()
	if v6 != nil {
		return ipl.lookup(v6)
	}
	return
}

func (ipl *IPList) lookup(ip net.IP) (Range, bool) {
	for _, r := range ipl.ranges {
		if len(r.First) != len(ip) || len(r.Last) != len(ip) {
			continue
		}
		if bytes.Compare(r.First, ip) <= 0 && bytes.Compare(ip, r.Last) <= 0 {
			return r, true
		}
	}
	return Range{}, false
}

func minifyIP(ip *net.IP) {
	if ip == nil || *ip == nil {
		return
	}
	v4 := ip.To4()
	if v4 != nil {
		*ip = append(make([]byte, 0, 4), v4...)
	}
}
