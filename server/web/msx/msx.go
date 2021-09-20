package msx

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	sets "server/settings"
	"server/torr"
	"server/torr/state"
	"server/utils"
	"server/version"

	"github.com/gin-gonic/gin"
)

type msxMenu struct {
	Logo      string        `json:"logo,omitempty"`
	Reuse     bool          `json:"reuse"`
	Cache     bool          `json:"cache"`
	Restore   bool          `json:"restore"`
	Reference string        `json:"reference,omitempty"`
	Menu      []msxMenuItem `json:"menu"`
}

type msxMenuItem struct {
	Icon  string  `json:"icon,omitempty"`
	Label string  `json:"label,omitempty"`
	Data  msxData `json:"data,omitempty"`
}

type msxTemplate struct {
	Type       string `json:"type,omitempty"`
	Layout     string `json:"layout,omitempty"`
	Color      string `json:"color,omitempty"`
	Icon       string `json:"icon,omitempty"`
	IconSize   string `json:"iconSize,omitempty"`
	BadgeColor string `json:"badgeColor,omitempty"`
	TagColor   string `json:"tagColor,omitempty"`
	Properties gin.H  `json:"properties,omitempty"`
}

type msxData struct {
	Type     string      `json:"type,omitempty"`
	Headline string      `json:"headline,omitempty"`
	Action   string      `json:"action,omitempty"`
	Template msxTemplate `json:"template,omitempty"`
	Items    []msxItem   `json:"items,omitempty"`
	Pages    []msxPage   `json:"pages,omitempty"`
}

type msxItem struct {
	Title       string `json:"title,omitempty"`
	Label       string `json:"label,omitempty"`
	PlayerLabel string `json:"playerLabel,omitempty"`
	Action      string `json:"action,omitempty"`
	Image       string `json:"image,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Badge       string `json:"badge,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Data        gin.H  `json:"data,omitempty"`
}

type msxPage struct {
	Items []gin.H `json:"items,omitempty"`
}

func msxStart(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":      "TorrServer",
		"version":   version.Version,
		"parameter": "menu:{PREFIX}{SERVER}/msx/torrents",
	})
}

// /msx/torrents
func msxTorrents(c *gin.Context) {
	torrs := torr.ListTorrent()

	host := utils.GetScheme(c) + "://" + c.Request.Host
	logo := host + "/apple-touch-icon.png"
	list := make([]msxItem, len(torrs))

	for i, tor := range torrs {
		item := msxItem{
			Title: tor.Title,
			Image: tor.Poster,
			Action: "content:" + host + "/msx/playlist/" + url.PathEscape(tor.Title) +
				"?hash=" + tor.TorrentSpec.InfoHash.HexString() + "&platform={PLATFORM}",
		}
		list[i] = item
	}

	c.JSON(200, msxMenu{
		Logo:      logo,
		Cache:     false,
		Reuse:     false,
		Restore:   false,
		Reference: host + "/msx/torrents",
		Menu: []msxMenuItem{
			// Main page
			{
				Icon:  "list",
				Label: "Torrents",
				Data: msxData{
					Type: "pages",
					Template: msxTemplate{
						Type:   "separate",
						Layout: "0,0,2,4",
						Icon:   "msx-white-soft:movie",
						Color:  "msx-glass",
					},
					Items: list,
				},
				// About
			}, {
				Icon:  "info",
				Label: "About",
				Data: msxData{
					Pages: []msxPage{
						{
							Items: []gin.H{
								{
									"type":        "default",
									"headline":    "TorrServer " + version.Version,
									"text":        "https://github.com/YouROK/TorrServer",
									"image":       logo,
									"imageFiller": "height-left",
									"imageWidth":  2,
									"layout":      "0,0,8,2",
									"color":       "msx-gray-soft",
								},
							},
						},
					},
				},
			},
		},
	})
}

// /msx/playlist?hash=...
func msxPlaylist(c *gin.Context) {
	hash, _ := c.GetQuery("hash")
	if hash == "" {
		c.JSON(200, msxData{
			Action: "error:Item not found",
		})
		return
	}
	platform, _ := c.GetQuery("platform")

	tor := torr.GetTorrent(hash)
	if tor == nil {
		c.JSON(200, msxData{
			Action: "error:Item not found",
		})
		return
	}

	if tor.Stat == state.TorrentInDB {
		tor = torr.LoadTorrent(tor)
		if tor == nil {
			c.JSON(200, msxData{
				Action: "error:Error while getting torrent info",
			})
			return
		}
	}

	host := utils.GetScheme(c) + "://" + c.Request.Host
	status := tor.Status()
	viewed := sets.ListViewed(hash)
	var list []msxItem

	for _, f := range status.FileStats {
		mime := utils.GetMimeType(f.Path)
		action := mime[0 : len(mime)-2]

		if action == "*" {
			continue
		}
		name := filepath.Base(f.Path)
		uri := host + "/stream/" + url.PathEscape(name) + "?link=" + hash + "&index=" + fmt.Sprint(f.Id) + "&play"
		item := msxItem{
			Label:       name,
			PlayerLabel: strings.TrimSuffix(name, filepath.Ext(name)),
			Action:      action + ":" + uri,
		}

		if platform == "android" || platform == "firetv" {
			item.Action = "system:tvx:launch"
			item.Data = gin.H{
				"id":   hash + "-" + fmt.Sprint(f.Id),
				"uri":  uri,
				"type": mime,
			}
		} else if platform == "lg" {
			// TODO - custom player needed
			//item.Action = "system:lg:launch:com.webos.app.mediadiscovery"
			//item.Data = gin.H{
			//	"properties": gin.H{
			//		"videoList": gin.H{
			//			"result": [1]gin.H{
			//				gin.H{
			//					"url":       uri,
			//					"thumbnail": tor.Poster,
			//				},
			//			},
			//		},
			//	},
			//}
		} else if platform == "ios" || platform == "mac" {
			// TODO - for iOS and Mac the application must be defined in scheme but we don't know what user has installed
			item.Action = "system:tvx:launch:vlc://" + uri
		}

		if isViewed(viewed, f.Id) {
			item.Tag = " "
		}
		if action == "audio" {
			item.Icon = "msx-white-soft:music-note"
		}
		list = append(list, item)
	}

	if len(list) == 0 {
		c.JSON(200, msxData{
			Action: "error:No supported content found",
		})
		return
	}

	res := msxData{
		Headline: tor.Title,
		Type:     "list",
		Template: msxTemplate{
			Type:       "control",
			Layout:     "0,2,12,1",
			Color:      "msx-glass",
			Icon:       "msx-white-soft:movie",
			IconSize:   "medium",
			BadgeColor: "msx-yellow",
			TagColor:   "msx-yellow",
		},
		Items: list,
	}

	if platform == "tizen" {
		res.Template.Properties = gin.H{
			"button:content:icon":   "tune",
			"button:content:action": "content:request:interaction:init@" + host + "/msx/tizen.html",
		}
	} else if platform == "netcast" {
		res.Template.Properties = gin.H{
			"button:content:icon":   "tune",
			"button:content:action": "system:netcast:menu",
		}
	}

	// If only one item start to play immediately but it not works
	// if (len(list) == 1) {
	//  res.Action = "execute:" + list[0].Action
	// }

	c.JSON(200, res)
}

func isViewed(viewed []*sets.Viewed, id int) bool {
	for _, v := range viewed {
		if v.FileIndex == id {
			return true
		}
	}
	return false
}
