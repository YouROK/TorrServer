package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"server/log"
)

const (
	baseURL      = "https://api.themoviedb.org/3"
	imageBaseURL = "https://image.tmdb.org/t/p/w500"
)

type SearchResult struct {
	Page         int      `json:"page"`
	Results      []Media  `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type Media struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`        // For movies
	Name         string  `json:"name"`         // For TV shows
	OriginalName string  `json:"original_name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`   // For movies
	FirstAirDate string  `json:"first_air_date"` // For TV shows
	VoteAverage  float64 `json:"vote_average"`
	MediaType    string  `json:"media_type"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ParseTorrentName extracts title, year, and season from torrent name
func ParseTorrentName(name string) (title string, year string, season string, isSeries bool) {
	// Replace common separators with spaces first
	cleaned := strings.ReplaceAll(name, ".", " ")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")
	
	// Check for season BEFORE removing quality tags
	seasonMatch := regexp.MustCompile(`(?i)\s+s(\d+)`).FindStringSubmatch(cleaned)
	if len(seasonMatch) > 1 {
		isSeries = true
		season = seasonMatch[1]
	}
	
	// Extract year BEFORE removing it
	yearMatch := regexp.MustCompile(`\s+(19\d{2}|20\d{2})\s+`).FindStringSubmatch(cleaned)
	if len(yearMatch) > 1 {
		year = yearMatch[1]
	}
	
	// Remove quality indicators and everything after them
	cleaned = regexp.MustCompile(`(?i)\s+(1080p|720p|2160p|4k|bluray|web-dl|web|webrip|hdtv|dvdrip|brrip|x264|x265|hevc|aac|ac3|dts).*`).ReplaceAllString(cleaned, "")
	
	// Remove season/episode markers from title
	cleaned = regexp.MustCompile(`(?i)\s+s\d+.*`).ReplaceAllString(cleaned, "")
	
	// Remove year from title
	if year != "" {
		cleaned = strings.ReplaceAll(cleaned, year, "")
	}
	
	// Clean up multiple spaces and trim
	title = strings.TrimSpace(cleaned)
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	
	return
}

// SearchMovie searches for a movie by title
func (c *Client) SearchMovie(title string, year string) (*Media, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}
	
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("query", title)
	if year != "" {
		params.Add("year", year)
	}
	
	searchURL := fmt.Sprintf("%s/search/movie?%s", baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search movie: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results found for: %s", title)
	}
	
	// Return first result
	media := &result.Results[0]
	media.MediaType = "movie"
	return media, nil
}

// SearchMulti searches for both movies and TV shows
func (c *Client) SearchMulti(title string) (*Media, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}
	
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("query", title)
	
	searchURL := fmt.Sprintf("%s/search/multi?%s", baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results found for: %s", title)
	}
	
	// Return first result (movie or TV)
	media := &result.Results[0]
	if media.MediaType == "" {
		// Determine type from fields
		if media.Title != "" {
			media.MediaType = "movie"
		} else {
			media.MediaType = "tv"
		}
	}
	return media, nil
}

// SearchTV searches for a TV show by title
func (c *Client) SearchTV(title string, year string) (*Media, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}
	
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("query", title)
	if year != "" {
		params.Add("first_air_date_year", year)
	}
	
	searchURL := fmt.Sprintf("%s/search/tv?%s", baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search TV show: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results found for: %s", title)
	}
	
	// Return first result
	media := &result.Results[0]
	media.MediaType = "tv"
	return media, nil
}

// SearchAuto automatically determines if it's a movie or TV show and searches
func (c *Client) SearchAuto(torrentName string) (*Media, error) {
	title, year, _, isSeries := ParseTorrentName(torrentName)
	
	log.TLogln("TMDB: Parsed torrent name:", torrentName)
	log.TLogln("TMDB: Title:", title, "Year:", year, "IsSeries:", isSeries)
	
	var media *Media
	var err error
	
	// Try multi-search first (works for both movies and TV shows)
	media, err = c.SearchMulti(title)
	
	// If not found and title contains Cyrillic, try extracting English from original
	if err != nil && containsCyrillic(title) {
		log.TLogln("TMDB: Russian title not found, trying to extract English from torrent name...")
		englishTitle := extractEnglishTitle(torrentName)
		if englishTitle != "" && englishTitle != title {
			log.TLogln("TMDB: Trying English title:", englishTitle)
			media, err = c.SearchMulti(englishTitle)
		}
	}
	
	if err != nil {
		log.TLogln("TMDB: Search failed:", err)
		return nil, err
	}
	
	log.TLogln("TMDB: Found:", media.GetTitle(), "Type:", media.MediaType, "ID:", media.ID)
	return media, nil
}

// containsCyrillic checks if string contains Cyrillic characters
func containsCyrillic(s string) bool {
	for _, r := range s {
		if (r >= 0x0400 && r <= 0x04FF) || (r >= 0x0500 && r <= 0x052F) {
			return true
		}
	}
	return false
}

// extractEnglishTitle tries to extract English title from torrent name
// Format: "Russian Name / English Name / ..."
func extractEnglishTitle(name string) string {
	// Common pattern: "Русское / English / Season: X"
	parts := strings.Split(name, "/")
	if len(parts) >= 2 {
		// Second part is usually English
		englishPart := strings.TrimSpace(parts[1])
		// Parse it same way as main title
		title, _, _, _ := ParseTorrentName(englishPart)
		return title
	}
	return ""
}

// GetTitle returns the appropriate title (for movie or TV show)
func (m *Media) GetTitle() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

// GetPosterURL returns the full poster URL
func (m *Media) GetPosterURL() string {
	if m.PosterPath == "" {
		return ""
	}
	return imageBaseURL + m.PosterPath
}

// GetBackdropURL returns the full backdrop URL
func (m *Media) GetBackdropURL() string {
	if m.BackdropPath == "" {
		return ""
	}
	return imageBaseURL + m.BackdropPath
}

// GetYear returns the release year
func (m *Media) GetYear() string {
	if m.ReleaseDate != "" {
		return strings.Split(m.ReleaseDate, "-")[0]
	}
	if m.FirstAirDate != "" {
		return strings.Split(m.FirstAirDate, "-")[0]
	}
	return ""
}
