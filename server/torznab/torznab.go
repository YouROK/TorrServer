package torznab

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"server/log"
	"server/rutor/models"
	"server/settings"
	"server/version"
)

type TorznabAttribute struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type TorznabEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type TorznabItem struct {
	Title       string             `xml:"title"`
	Link        string             `xml:"link"`
	Description string             `xml:"description"`
	PubDate     string             `xml:"pubDate"`
	Size        int64              `xml:"size"`
	Indexer     string             `xml:"jackettindexer"`
	Prowlarr    string             `xml:"prowlarrindexer"`
	Enclosure   []TorznabEnclosure `xml:"enclosure"`
	Attributes  []TorznabAttribute `xml:"attr"`
}

type TorznabChannel struct {
	Items []TorznabItem `xml:"item"`
}

type TorznabResponse struct {
	Channel TorznabChannel `xml:"channel"`
}

// httpGet performs a GET request with a browser-like User-Agent, because
// some Torznab indexers (e.g. behind a WAF) return 403 without one.
func httpGet(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TorrServer-Torznab/"+version.Version)
	return http.DefaultClient.Do(req)
}

// buildAPIURL normalizes a Torznab host and returns the parsed URL
// pointing to the "api" endpoint. It accepts hosts with or without an
// http(s):// scheme and with or without a trailing "/api" or "/api/" segment.
func buildAPIURL(host string) (*url.URL, error) {
	host = strings.TrimSpace(host)
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	host = strings.TrimRight(host, "/")
	host = strings.TrimSuffix(host, "/api")
	host = strings.TrimRight(host, "/")

	return url.Parse(host + "/api")
}

func Search(query string, index int) []*models.TorrentDetails {
	if !settings.BTsets.EnableTorznabSearch || len(settings.BTsets.TorznabUrls) == 0 {
		return nil
	}

	var allResults []*models.TorrentDetails
	if index >= 0 && index < len(settings.BTsets.TorznabUrls) {
		config := settings.BTsets.TorznabUrls[index]
		if config.Host != "" && config.Key != "" {
			return searchOne(
				config.Host,
				config.Key,
				config.Categories,
				config.CatType,
				query,
			)
		}
		return nil
	}

	for _, config := range settings.BTsets.TorznabUrls {
		if config.Host == "" || config.Key == "" {
			continue
		}
		results := searchOne(
			config.Host,
			config.Key,
			config.Categories,
			config.CatType,
			query,
		)
		if results != nil {
			allResults = append(allResults, results...)
		}
	}
	return allResults
}

func searchOne(host, key, categories string, categoryType settings.CategoryType, query string) []*models.TorrentDetails {
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	u, err := buildAPIURL(host)
	if err != nil {
		log.TLogln("Error parsing Torznab host:", err)
		return nil
	}

	q := u.Query()
	q.Set("apikey", key)
	q.Set("t", "search")
	q.Set("q", query)

	switch categoryType {
	case settings.CategoryAll:
	case settings.CategoryManual:
		if categories != "" {
			q.Set("cat", categories)
		}
	default:
		q.Set("cat", "5000,2000") // Movies, TV
	}

	u.RawQuery = q.Encode()

	resp, err := httpGet(u.String())
	if err != nil {
		log.TLogln("Error connecting to Torznab:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.TLogln("Torznab returned status:", resp.Status)
		return nil
	}

	var torznabResp TorznabResponse
	if err := xml.NewDecoder(resp.Body).Decode(&torznabResp); err != nil {
		log.TLogln("Error decoding Torznab response:", err)
		return nil
	}

	var results []*models.TorrentDetails
	for _, item := range torznabResp.Channel.Items {
		if item.Indexer == "" {
			item.Indexer = item.Prowlarr
		}
		detail := &models.TorrentDetails{
			Title:      item.Title,
			Name:       item.Title, // Use Title as Name for now
			Link:       item.Link,
			Tracker:    item.Indexer,
			CreateDate: parseDate(item.PubDate),
		}

		if len(item.Enclosure) > 0 {
			detail.Link = item.Enclosure[0].URL
			detail.Size = formatSize(item.Enclosure[0].Length)
		} else {
			detail.Size = formatSize(item.Size)
		}

		for _, attr := range item.Attributes {
			if attr.Name == "magneturl" {
				detail.Magnet = attr.Value
				detail.Hash = extractHash(detail.Magnet)
			}
			if attr.Name == "seeders" {
				detail.Seed, _ = strconv.Atoi(attr.Value)
			}
			if attr.Name == "peers" {
				detail.Peer, _ = strconv.Atoi(attr.Value)
			}
		}

		// Fallback if magnet not in attributes but link is a magnet
		if detail.Magnet == "" && strings.HasPrefix(detail.Link, "magnet:") {
			detail.Magnet = detail.Link
			detail.Hash = extractHash(detail.Magnet)
		}

		results = append(results, detail)
	}

	return results
}

func Test(host, key string) error {
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	u, err := buildAPIURL(host)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("apikey", key)
	q.Set("t", "caps")
	u.RawQuery = q.Encode()

	resp, err := httpGet(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s", resp.Status)
	}

	var probe struct {
		XMLName     xml.Name
		Code        string `xml:"code,attr"`
		Description string `xml:"description,attr"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&probe); err != nil {
		return fmt.Errorf("invalid xml response: %v", err)
	}

	if probe.XMLName.Local == "error" {
		msg := probe.Description
		if msg == "" {
			msg = probe.Code
		}
		return fmt.Errorf("api error: %s", msg)
	}

	if probe.XMLName.Local != "caps" {
		return fmt.Errorf("unexpected xml root: %s", probe.XMLName.Local)
	}

	return nil
}

func parseDate(dateStr string) time.Time {
	// RFC1123 is common in RSS
	t, err := time.Parse(time.RFC1123, dateStr)
	if err != nil {
		// Try RFC1123Z
		t, err = time.Parse(time.RFC1123Z, dateStr)
		if err != nil {
			return time.Now()
		}
	}
	return t
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func extractHash(magnet string) string {
	if strings.HasPrefix(magnet, "magnet:?") {
		u, err := url.Parse(magnet)
		if err == nil {
			xt := u.Query().Get("xt")
			if strings.HasPrefix(xt, "urn:btih:") {
				return strings.TrimPrefix(xt, "urn:btih:")
			}
		}
	}
	return ""
}
