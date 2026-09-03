package waf

import (
	"bufio"
	"net"
	"net/url"
	"strings"
	"unicode"
)

// defaultBlockedReferers is always enforced for Referer/Origin checks.
// These hosts cannot be disabled via settings and are not bypassed by the IP whitelist.
var defaultBlockedReferers = []string{
	"abhq.ru",
	"abmsx.tech",
	"akter.black",
	"bylampa.online",
	"lampa.click",
	"lampa.land",
	"lampa1.ru",
	"line.pm",
	"nnmtv.pw",
	"stull.xyz",
	"tvigl.info",
	"uspeh.sbs",
	"usph.xyz",
	"xabb.ru",
}

func blockedReferersFromConfig(buf []byte) []string {
	return mergeReferers(defaultBlockedReferers, scanRefererBuf(buf))
}

func mergeReferers(defaults, fileHosts []string) []string {
	seen := make(map[string]struct{}, len(defaults)+len(fileHosts))
	blocked := make([]string, 0, len(defaults)+len(fileHosts))
	for _, host := range defaults {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		blocked = append(blocked, host)
	}
	for _, host := range fileHosts {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		blocked = append(blocked, host)
	}
	return blocked
}

func scanRefererBuf(buf []byte) []string {
	hosts, _ := scanRefererBufValidated(buf)
	return hosts
}

func scanRefererBufValidated(buf []byte) ([]string, []ParseWarning) {
	if len(buf) == 0 {
		return nil, nil
	}
	var hosts []string
	var warnings []ParseWarning
	scanner := bufio.NewScanner(strings.NewReader(string(buf)))
	scanner.Buffer(make([]byte, 64*1024), maxScannerTokenSize)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		host, ok := normalizeRefererRule(line)
		if !ok {
			warnings = append(warnings, ParseWarning{Line: lineNo, Code: "invalid_referer"})
			continue
		}
		hosts = append(hosts, host)
	}
	if scanner.Err() != nil {
		warnings = append(warnings, ParseWarning{Line: lineNo + 1, Code: "line_too_long"})
	}
	return hosts, warnings
}

func normalizeRefererRule(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "*") {
		return "", false
	}

	host := value
	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") ||
			u.User != nil || (u.Path != "" && u.Path != "/") ||
			u.RawQuery != "" || u.Fragment != "" {
			return "", false
		}
		host = u.Hostname()
	} else {
		if strings.ContainsAny(value, "/?#@") {
			return "", false
		}
		if parsedHost, _, err := net.SplitHostPort(value); err == nil {
			host = parsedHost
		}
	}

	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	host = strings.Trim(host, "[]")
	if host == "" {
		return "", false
	}
	if net.ParseIP(host) != nil {
		return host, true
	}
	if len(host) > 253 {
		return "", false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, r := range label {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' {
				return "", false
			}
		}
	}
	return host, true
}

func hostFromHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	u, err := url.Parse(value)
	if err != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "" {
		return ""
	}
	return host
}

func hostMatchesBlocked(host string, blocked []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	for _, blockedHost := range blocked {
		if blockedHost == "" {
			continue
		}
		if host == blockedHost || strings.HasSuffix(host, "."+blockedHost) {
			return true
		}
	}
	return false
}

func isBlockedReferer(referer string, origin string, blocked []string) (blockedHost string, ok bool) {
	if host := hostFromHeader(referer); host != "" {
		if hostMatchesBlocked(host, blocked) {
			return host, true
		}
	}
	if host := hostFromHeader(origin); host != "" {
		if hostMatchesBlocked(host, blocked) {
			return host, true
		}
	}
	return "", false
}
