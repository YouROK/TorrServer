package api

import (
  "fmt"
  "strings"
  "net/url"
  "path/filepath"

  sets "server/settings"
  "server/torr"
  "server/torr/state"
  "server/utils"
  "server/version"

  "github.com/gin-gonic/gin"
)

type msxMenu struct {
  Logo string `json:"logo,omitempty"`
  Menu []msxMenuItem `json:"menu"`
}

type msxMenuItem struct {
  Icon string `json:"icon,omitempty"`
  Label string `json:"label,omitempty"`
  Data msxData `json:"data,omitempty"`
}

type msxData struct {
  Type string `json:"type,omitempty"`
  Headline string `json:"headline,omitempty"`
  Action string `json:"action,omitempty"`
  Template gin.H `json:"template,omitempty"`
  Items []msxItem `json:"items,omitempty"`
  Pages []msxPage `json:"pages,omitempty"`
}

type msxItem struct {
  Title string `json:"title,omitempty"`
  Label string `json:"label,omitempty"`
  PlayerLabel string `json:"playerLabel,omitempty"`
  Action string `json:"action,omitempty"`
  Image string `json:"image,omitempty"`
  Icon string `json:"icon,omitempty"`
  Badge string `json:"badge,omitempty"`
  Tag string `json:"tag,omitempty"`
}

type msxPage struct {
  Items []gin.H `json:"items,omitempty"`
}

func msxStart(c *gin.Context) {
  c.JSON(200, gin.H{
    "name": "TorrServer",
    "version": version.Version,
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
      Action: "content:" + host + "/msx/playlist/" + url.PathEscape(tor.Title) + "?hash=" + tor.TorrentSpec.InfoHash.HexString(),
    }
    list[i] = item
  }

  c.JSON(200, msxMenu{
    Logo: logo,
    Menu: []msxMenuItem{
      // Main page
      {
        Icon: "list",
        Label: "Torrents",
        Data: msxData{
          Type: "pages",
          Template: gin.H{
            "type": "separate",
            "layout": "0,0,2,4",
            "icon": "msx-white-soft:movie",
            "color": "msx-glass",
          },
          Items: list,
        },
      // About
      },{
        Icon: "info",
        Label: "About",
        Data: msxData{
          Pages: []msxPage{
            {
              Items: []gin.H{
                {
                  "type": "default",
                  "headline": "TorrServer " + version.Version,
                  "text": "https://github.com/YouROK/TorrServer",
                  "image": logo,
                  "imageFiller": "height-left",
                  "imageWidth": 2,
                  "layout": "0,0,8,2",
                  "color": "msx-gray-soft",
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
  list := []msxItem{}

  for _, f := range status.FileStats {
    mime := utils.GetMimeType(f.Path)
    action := mime[0 : len(mime)-2]

    if (action == "*") {
      continue
    }
    name := filepath.Base(f.Path)
    item := msxItem{
      Label: name,
      PlayerLabel: strings.TrimSuffix(name, filepath.Ext(name)),
      Action: action + ":" + host + "/stream/" + url.PathEscape(name) + "?link=" + hash + "&index=" + fmt.Sprint(f.Id) + "&play",
    }
    if (isViewed(viewed, f.Id)) {
      item.Tag = " "
    }
    if (action == "audio") {
      item.Icon = "msx-white-soft:music-note"
    }
    list = append(list, item)
  }

  if (len(list) == 0) {
      c.JSON(200, msxData{
        Action: "error:No supported content found",
      })
      return
  }

  res := msxData{
    Headline: tor.Title,
    Type: "list",
    Template: gin.H{
      "type": "control",
      "layout": "0,2,12,1",
      "color": "msx-glass",
      "icon": "msx-white-soft:movie",
      "iconSize": "medium",
      "badgeColor": "msx-yellow",
      "tagColor": "msx-yellow",
    },
    Items: list,
  }

  // If only one item start to play immediately but it not works
  // if (len(list) == 1) {
  //  res.Action = "execute:" + list[0].Action
  // }

  c.JSON(200, res)
}

func isViewed(viewed []*sets.Viewed, id int) bool {
  for _, v := range viewed {
    if (v.FileIndex == id) {
      return true
    }
  }
  return false
}
