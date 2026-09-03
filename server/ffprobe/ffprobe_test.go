package ffprobe

import (
	"encoding/json"
	"math"
	"testing"

	goffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

// GET /ffp used to return HTTP 400 for HDR media because go-ffprobe v2.2.1
// expected integer mastering-display fields while ffprobe emits rationals.
func TestProbeDataUnmarshalsRationalHDRMasteringDisplay(t *testing.T) {
	raw := []byte(`{
		"format": {
			"bit_rate": "12345678",
			"duration": "3600.000000",
			"size": "5550000000"
		},
		"streams": [
			{
				"codec_type": "video",
				"side_data_list": [
					{
						"side_data_type": "Mastering display metadata",
						"red_x": "17/25",
						"red_y": "8/25",
						"max_luminance": "4000/1"
					}
				]
			}
		]
	}`)

	var data goffprobe.ProbeData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal HDR ffprobe JSON: %v", err)
	}

	if data.Format == nil {
		t.Fatal("expected format")
	}
	if data.Format.BitRate != "12345678" {
		t.Fatalf("BitRate = %q", data.Format.BitRate)
	}
	if data.Format.DurationSeconds != 3600 {
		t.Fatalf("DurationSeconds = %v", data.Format.DurationSeconds)
	}
	if data.Format.Size != "5550000000" {
		t.Fatalf("Size = %q", data.Format.Size)
	}

	video := data.FirstVideoStream()
	if video == nil {
		t.Fatal("expected video stream")
	}

	mdm, err := video.SideDataList.GetMasteringDisplayMetadata()
	if err != nil {
		t.Fatalf("GetMasteringDisplayMetadata: %v", err)
	}
	assertFlexFloat(t, "red_x", mdm.RedX, 17.0/25.0)
	assertFlexFloat(t, "red_y", mdm.RedY, 8.0/25.0)
	assertFlexFloat(t, "max_luminance", mdm.MaxLuminance, 4000)
}

func TestProbeDataUnmarshalsIntegerHDRMasteringDisplay(t *testing.T) {
	raw := []byte(`{
		"streams": [
			{
				"codec_type": "video",
				"side_data_list": [
					{
						"side_data_type": "Mastering display metadata",
						"red_x": 34000,
						"red_y": 16000,
						"max_luminance": 1000
					}
				]
			}
		]
	}`)

	var data goffprobe.ProbeData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal integer HDR ffprobe JSON: %v", err)
	}

	video := data.FirstVideoStream()
	if video == nil {
		t.Fatal("expected video stream")
	}

	mdm, err := video.SideDataList.GetMasteringDisplayMetadata()
	if err != nil {
		t.Fatalf("GetMasteringDisplayMetadata: %v", err)
	}
	assertFlexFloat(t, "red_x", mdm.RedX, 34000)
	assertFlexFloat(t, "red_y", mdm.RedY, 16000)
	assertFlexFloat(t, "max_luminance", mdm.MaxLuminance, 1000)
}

func assertFlexFloat(t *testing.T, name string, got goffprobe.FlexFloat, want float64) {
	t.Helper()
	if math.Abs(float64(got)-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
