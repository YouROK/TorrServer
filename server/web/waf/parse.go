package waf

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strings"
)

const maxScannerTokenSize = 1024 * 1024

type ParseWarning struct {
	Line int
	Code string
}

func scanBuf(buf []byte) (Ranger, []ParseWarning) {
	if len(buf) == 0 {
		return New(nil), nil
	}
	var ranges []Range
	var warnings []ParseWarning
	scanner := bufio.NewScanner(strings.NewReader(string(buf)))
	scanner.Buffer(make([]byte, 64*1024), maxScannerTokenSize)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		r, ok, err := parseLine(scanner.Bytes())
		if err != nil {
			warnings = append(warnings, ParseWarning{Line: lineNo, Code: "invalid_ip_range"})
			continue
		}
		if ok {
			ranges = append(ranges, r)
		}
	}
	if err := scanner.Err(); err != nil {
		warnings = append(warnings, ParseWarning{Line: lineNo + 1, Code: "line_too_long"})
	}
	return New(ranges), warnings
}

// parseLine parses one IP list line.
// Supported forms:
//   - # comment / blank
//   - ip
//   - first-last
//   - cidr (e.g. 10.0.0.0/8, 2001:db8::/32)
//   - description:ip|first-last|cidr  (description is text before the first ':')
func parseLine(l []byte) (r Range, ok bool, err error) {
	l = bytes.TrimSpace(l)
	if len(l) == 0 || bytes.HasPrefix(l, []byte("#")) {
		return
	}
	line := string(l)

	first, last, perr := parseIPSpec(line)
	if perr == nil {
		r.First, r.Last = first, last
		ok = true
		return
	}

	colon := strings.IndexByte(line, ':')
	if colon <= 0 {
		err = errors.New("bad IP range")
		return
	}
	desc := strings.TrimSpace(line[:colon])
	spec := strings.TrimSpace(line[colon+1:])
	if desc == "" || spec == "" {
		err = errors.New("bad IP range")
		return
	}
	first, last, perr = parseIPSpec(spec)
	if perr != nil {
		err = perr
		return
	}
	r.Description = desc
	r.First, r.Last = first, last
	ok = true
	return
}

func parseIPSpec(spec string) (first, last net.IP, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil, errors.New("bad IP range")
	}

	if strings.Contains(spec, "/") {
		_, network, cerr := net.ParseCIDR(spec)
		if cerr != nil {
			return nil, nil, errors.New("bad IP range")
		}
		first = append(net.IP(nil), network.IP...)
		minifyIP(&first)
		last = lastIP(network)
		minifyIP(&last)
		if first == nil || last == nil || len(first) != len(last) {
			return nil, nil, errors.New("bad IP range")
		}
		return first, last, nil
	}

	if hyphen := findRangeHyphen(spec); hyphen >= 0 {
		left := strings.TrimSpace(spec[:hyphen])
		right := strings.TrimSpace(spec[hyphen+1:])
		first = net.ParseIP(left)
		last = net.ParseIP(right)
		minifyIP(&first)
		minifyIP(&last)
		if first == nil || last == nil || len(first) != len(last) {
			return nil, nil, errors.New("bad IP range")
		}
		if bytes.Compare(first, last) > 0 {
			return nil, nil, errors.New("bad IP range")
		}
		return first, last, nil
	}

	first = net.ParseIP(spec)
	minifyIP(&first)
	if first == nil {
		return nil, nil, errors.New("bad IP range")
	}
	last = first
	return first, last, nil
}

// findRangeHyphen finds a hyphen that splits spec into two valid IP addresses.
// Prefer the first successful split from the left after at least one character.
func findRangeHyphen(spec string) int {
	for i := 1; i < len(spec)-1; i++ {
		if spec[i] != '-' {
			continue
		}
		left := strings.TrimSpace(spec[:i])
		right := strings.TrimSpace(spec[i+1:])
		a := net.ParseIP(left)
		b := net.ParseIP(right)
		if a != nil && b != nil {
			return i
		}
	}
	return -1
}

func lastIP(network *net.IPNet) net.IP {
	ip := append(net.IP(nil), network.IP...)
	for i := range ip {
		ip[i] |= ^network.Mask[i]
	}
	return ip
}
