package playback

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	configFileName = "playback_devices.json"
	configVersion  = 1

	RoutingLocal      = "local"
	RoutingPrimary    = "primary"
	RoutingPerBrowser = "per-browser"
)

var (
	idPattern                   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	ErrRemotePlaybackDisabled   = errors.New("remote playback is disabled")
	ErrPrimaryDeviceNotSelected = errors.New("primary playback device is not selected")
)

// Device is the private, persisted playback-device configuration.
// Endpoint and Token are intentionally never exposed by PublicConfig.
type Device struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	Token         string `json:"token,omitempty"`
	StreamBaseURL string `json:"stream_base_url,omitempty"`
	Fullscreen    bool   `json:"fullscreen,omitempty"`
}

// PublicDevice is safe to return to normal web clients.
type PublicDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ManagedDevice is returned only by the authenticated management endpoint.
// The token itself is never returned.
type ManagedDevice struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	StreamBaseURL string `json:"stream_base_url,omitempty"`
	Fullscreen    bool   `json:"fullscreen"`
	HasToken      bool   `json:"has_token"`
}

// DeviceInput is used for create/update operations. An empty token preserves
// the current token. ClearToken explicitly removes it when no replacement is supplied.
type DeviceInput struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	Token         string `json:"token,omitempty"`
	ClearToken    bool   `json:"clear_token,omitempty"`
	StreamBaseURL string `json:"stream_base_url,omitempty"`
	Fullscreen    bool   `json:"fullscreen"`
}

// Settings controls whether the new routing behavior is enabled. Local routing
// preserves the historical behavior where each browser opens its own player.
type Settings struct {
	Enabled         bool   `json:"enabled"`
	RoutingMode     string `json:"routing_mode"`
	PrimaryDeviceID string `json:"primary_device_id,omitempty"`
}

type PublicConfig struct {
	Settings
	Devices []PublicDevice `json:"devices"`
}

type ManagedConfig struct {
	Settings
	Devices []ManagedDevice `json:"devices"`
}

type persistedConfig struct {
	Version int `json:"version"`
	Settings
	Devices []Device `json:"devices"`
}

// AgentPlayRequest is the contract sent by TorrServer to a playback agent.
type AgentPlayRequest struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	Index      int    `json:"index"`
	StreamURL  string `json:"stream_url"`
	Fullscreen bool   `json:"fullscreen"`
}

type store struct {
	mu       sync.RWMutex
	path     string
	readOnly bool
	settings Settings
	devices  map[string]Device
	client   *http.Client
}

var global = &store{
	settings: defaultSettings(),
	devices:  make(map[string]Device),
	client: &http.Client{
		Timeout: 12 * time.Second,
	},
}

func defaultSettings() Settings {
	return Settings{Enabled: false, RoutingMode: RoutingLocal}
}

// Init loads playback devices from the TorrServer configuration directory.
func Init(configDir string, readOnly bool) error {
	global.mu.Lock()
	defer global.mu.Unlock()

	global.path = filepath.Join(configDir, configFileName)
	global.readOnly = readOnly
	global.settings = defaultSettings()
	global.devices = make(map[string]Device)

	buf, err := os.ReadFile(global.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read playback devices: %w", err)
	}

	config, err := decodeConfig(buf)
	if err != nil {
		return err
	}
	settings, err := normalizeSettings(config.Settings, nil)
	if err != nil {
		return fmt.Errorf("invalid playback settings: %w", err)
	}

	for _, device := range config.Devices {
		normalized, err := normalizeDevice(device)
		if err != nil {
			return fmt.Errorf("invalid playback device %q: %w", device.ID, err)
		}
		global.devices[normalized.ID] = normalized
	}
	if settings.PrimaryDeviceID != "" {
		if _, ok := global.devices[settings.PrimaryDeviceID]; !ok {
			settings = defaultSettings()
		}
	}
	global.settings = settings
	return nil
}

func decodeConfig(buf []byte) (persistedConfig, error) {
	trimmed := strings.TrimSpace(string(buf))
	if strings.HasPrefix(trimmed, "[") {
		// Backward compatibility with the first prototype, which stored only an array.
		var devices []Device
		if err := json.Unmarshal(buf, &devices); err != nil {
			return persistedConfig{}, fmt.Errorf("decode playback devices: %w", err)
		}
		return persistedConfig{Version: configVersion, Settings: defaultSettings(), Devices: devices}, nil
	}

	var config persistedConfig
	if err := json.Unmarshal(buf, &config); err != nil {
		return persistedConfig{}, fmt.Errorf("decode playback devices: %w", err)
	}
	if config.Version == 0 {
		config.Version = configVersion
	}
	if config.RoutingMode == "" {
		config.RoutingMode = RoutingLocal
	}
	return config, nil
}

func GetPublicConfig() PublicConfig {
	global.mu.RLock()
	defer global.mu.RUnlock()
	return PublicConfig{Settings: global.settings, Devices: publicDevicesLocked()}
}

func GetManagedConfig() ManagedConfig {
	global.mu.RLock()
	defer global.mu.RUnlock()

	devices := make([]ManagedDevice, 0, len(global.devices))
	for _, device := range global.devices {
		devices = append(devices, ManagedDevice{
			ID:            device.ID,
			Name:          device.Name,
			Endpoint:      device.Endpoint,
			StreamBaseURL: device.StreamBaseURL,
			Fullscreen:    device.Fullscreen,
			HasToken:      device.Token != "",
		})
	}
	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
	})
	return ManagedConfig{Settings: global.settings, Devices: devices}
}

func publicDevicesLocked() []PublicDevice {
	devices := make([]PublicDevice, 0, len(global.devices))
	for _, device := range global.devices {
		devices = append(devices, PublicDevice{ID: device.ID, Name: device.Name})
	}
	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
	})
	return devices
}

// GetPrivate is intended only for server-side playback routing.
// Token is omitted from the returned copy.
func GetPrivate(id string) (Device, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	device, ok := global.devices[id]
	if !ok {
		return Device{}, false
	}
	device.Token = ""
	return device, true
}

func UpdateSettings(input Settings) (Settings, error) {
	global.mu.Lock()
	defer global.mu.Unlock()

	if global.readOnly {
		return Settings{}, errors.New("playback devices are read-only")
	}
	settings, err := normalizeSettings(input, global.devices)
	if err != nil {
		return Settings{}, err
	}
	previous := global.settings
	global.settings = settings
	if err := global.saveLocked(); err != nil {
		global.settings = previous
		return Settings{}, err
	}
	return global.settings, nil
}

func normalizeSettings(settings Settings, devices map[string]Device) (Settings, error) {
	settings.RoutingMode = strings.TrimSpace(settings.RoutingMode)
	settings.PrimaryDeviceID = strings.TrimSpace(settings.PrimaryDeviceID)
	if settings.RoutingMode == "" {
		settings.RoutingMode = RoutingLocal
	}
	switch settings.RoutingMode {
	case RoutingLocal, RoutingPrimary, RoutingPerBrowser:
	default:
		return Settings{}, errors.New("invalid playback routing mode")
	}
	if settings.RoutingMode == RoutingPrimary && settings.Enabled {
		if settings.PrimaryDeviceID == "" {
			return Settings{}, ErrPrimaryDeviceNotSelected
		}
		if devices != nil {
			if _, ok := devices[settings.PrimaryDeviceID]; !ok {
				return Settings{}, os.ErrNotExist
			}
		}
	}
	return settings, nil
}

func Upsert(input DeviceInput) (PublicDevice, error) {
	global.mu.Lock()
	defer global.mu.Unlock()

	if global.readOnly {
		return PublicDevice{}, errors.New("playback devices are read-only")
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		var err error
		id, err = randomID()
		if err != nil {
			return PublicDevice{}, err
		}
	}

	existing := global.devices[id]
	token := strings.TrimSpace(input.Token)
	if token == "" && !input.ClearToken {
		token = existing.Token
	}

	device, err := normalizeDevice(Device{
		ID:            id,
		Name:          input.Name,
		Endpoint:      input.Endpoint,
		Token:         token,
		StreamBaseURL: input.StreamBaseURL,
		Fullscreen:    input.Fullscreen,
	})
	if err != nil {
		return PublicDevice{}, err
	}

	global.devices[id] = device
	if err := global.saveLocked(); err != nil {
		if existing.ID == "" {
			delete(global.devices, id)
		} else {
			global.devices[id] = existing
		}
		return PublicDevice{}, err
	}

	return PublicDevice{ID: device.ID, Name: device.Name}, nil
}

func Delete(id string) error {
	global.mu.Lock()
	defer global.mu.Unlock()

	if global.readOnly {
		return errors.New("playback devices are read-only")
	}
	device, ok := global.devices[id]
	if !ok {
		return os.ErrNotExist
	}
	previousSettings := global.settings
	delete(global.devices, id)
	if global.settings.PrimaryDeviceID == id {
		global.settings.PrimaryDeviceID = ""
		global.settings.Enabled = false
		global.settings.RoutingMode = RoutingLocal
	}
	if err := global.saveLocked(); err != nil {
		global.devices[id] = device
		global.settings = previousSettings
		return err
	}
	return nil
}

// ResolveTarget enforces the server-wide routing mode. In primary mode the
// browser cannot override the configured main playback device.
func ResolveTarget(requestedID string) (string, error) {
	global.mu.RLock()
	defer global.mu.RUnlock()

	if !global.settings.Enabled || global.settings.RoutingMode == RoutingLocal {
		return "", ErrRemotePlaybackDisabled
	}
	if global.settings.RoutingMode == RoutingPrimary {
		if global.settings.PrimaryDeviceID == "" {
			return "", ErrPrimaryDeviceNotSelected
		}
		if _, ok := global.devices[global.settings.PrimaryDeviceID]; !ok {
			return "", os.ErrNotExist
		}
		return global.settings.PrimaryDeviceID, nil
	}
	requestedID = strings.TrimSpace(requestedID)
	if requestedID == "" {
		return "", errors.New("playback device is not selected")
	}
	if _, ok := global.devices[requestedID]; !ok {
		return "", os.ErrNotExist
	}
	return requestedID, nil
}

func Play(deviceID string, payload AgentPlayRequest) error {
	device, err := getDevice(deviceID)
	if err != nil {
		return err
	}
	payload.Fullscreen = device.Fullscreen
	return sendJSON(device, http.MethodPost, "play", payload)
}

func Test(deviceID string) error {
	device, err := getDevice(deviceID)
	if err != nil {
		return err
	}
	return sendJSON(device, http.MethodGet, "health", nil)
}

func getDevice(id string) (Device, error) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	device, ok := global.devices[id]
	if !ok {
		return Device{}, os.ErrNotExist
	}
	return device, nil
}

func sendJSON(device Device, method, endpointPath string, payload any) error {
	base, err := url.Parse(device.Endpoint)
	if err != nil {
		return err
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/" + endpointPath

	var body *strings.Reader
	if payload == nil {
		body = strings.NewReader("")
	} else {
		buf, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(buf))
	}

	req, err := http.NewRequest(method, base.String(), body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if device.Token != "" {
		req.Header.Set("Authorization", "Bearer "+device.Token)
	}

	resp, err := global.client.Do(req)
	if err != nil {
		return fmt.Errorf("playback device request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("playback device returned %s", resp.Status)
	}
	return nil
}

func normalizeDevice(device Device) (Device, error) {
	device.ID = strings.TrimSpace(device.ID)
	device.Name = strings.TrimSpace(device.Name)
	device.Endpoint = strings.TrimRight(strings.TrimSpace(device.Endpoint), "/")
	device.StreamBaseURL = strings.TrimRight(strings.TrimSpace(device.StreamBaseURL), "/")
	device.Token = strings.TrimSpace(device.Token)

	if !idPattern.MatchString(device.ID) {
		return Device{}, errors.New("invalid device id")
	}
	if device.Name == "" || len(device.Name) > 80 {
		return Device{}, errors.New("device name must contain 1-80 characters")
	}
	if len(device.Token) > 512 {
		return Device{}, errors.New("device token is too long")
	}
	if err := validateBaseURL(device.Endpoint, true); err != nil {
		return Device{}, fmt.Errorf("invalid endpoint: %w", err)
	}
	if err := validateBaseURL(device.StreamBaseURL, false); err != nil {
		return Device{}, fmt.Errorf("invalid stream base URL: %w", err)
	}
	return device, nil
}

func validateBaseURL(value string, required bool) error {
	if value == "" {
		if required {
			return errors.New("URL is required")
		}
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http and https are supported")
	}
	if parsed.Host == "" {
		return errors.New("host is required")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("credentials, query and fragment are not allowed")
	}
	return nil
}

func randomID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate device id: %w", err)
	}
	return "player-" + hex.EncodeToString(buf), nil
}

func (s *store) saveLocked() error {
	devices := make([]Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })

	buf, err := json.MarshalIndent(persistedConfig{
		Version:  configVersion,
		Settings: s.settings,
		Devices:  devices,
	}, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Chmod(s.path, 0o600)
}
