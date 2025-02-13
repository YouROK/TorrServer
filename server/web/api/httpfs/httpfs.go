package httpfs

import (
	"strings"
	"net/http"
	"net/url"
    "strconv"
	"server/log"
	"github.com/gin-gonic/gin"
	"server/torr"
	"github.com/pkg/errors"
	"github.com/anacrolix/torrent"

	"server/torr/state"

    sf "github.com/sa-/slicefunk"
)

func listDir(path string, folders [] string, files []string) string {
    result := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN\" \"http://www.w3.org/TR/html4/strict.dtd\">
               <html>
               <head>
               <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
               <title>Directory listing for ` + path + `</title>
                </head><body><h1>Directory listing for ` + path + `</h1><hr><ul>`

    for _, str := range folders {
        result += `<li><a href="` + url.PathEscape(str) + `/">` + str + `/</a></li>`
    }
    for _, str := range files {
        result += `<li><a href="` + url.PathEscape(str) + `">` + str + `</a></li>`
    }
    result += `</ul><hr></body></html>`
    return result
}



func HandleHttpfs(c *gin.Context) {
    path := c.Param("path")
    log.TLogln("URL path", path)
    path = strings.Trim(path, "/")

    if path == "" {
        newArray := sf.Map(torr.ListTorrent(), func(item * torr.Torrent) string { return item.Title })
        c.Header("Content-Type", "text/html; charset=utf-8")
        c.String(200, listDir(path, newArray, []string{}))
    } else {
        folders := strings.Split(path, "/")
        trName := folders[0]
        folderPath := strings.Join(folders[1:], "/")
//         for _, item := range torr.ListTorrent() {
//             log.TLogln("Torrents found", item.Title)
//         }
        requestedTorrents := sf.Filter(torr.ListTorrent(), func(item * torr.Torrent) bool { return item.Title == trName })
        if len(requestedTorrents) != 1 {
		    c.AbortWithError(http.StatusBadRequest, errors.New("Torrent not found" + trName + ": " + strconv.Itoa(len(requestedTorrents))))
		    return
        }
        requestedTorrent := torr.GetTorrent(requestedTorrents[0].Hash().HexString())
        if requestedTorrent == nil {
            c.AbortWithStatus(http.StatusNotFound)
            return
        }

        if requestedTorrent.Stat == state.TorrentInDB {
            requestedTorrent = torr.LoadTorrent(requestedTorrent)
            if requestedTorrent == nil {
                c.AbortWithError(http.StatusInternalServerError, errors.New("error get torrent info"))
                return
            }
        }

        newArray := sf.Filter(requestedTorrent.Files(), func(item * torrent.File) bool { return strings.HasPrefix(item.Path(), folderPath) })

//         for _, item := range newArray {
//             log.TLogln("Filtered files", item.Path(), " sha1: ")// + item.FileInfo())
//         }
//
        if len(newArray) == 1 && newArray[0].Path() == folderPath {
            log.TLogln("Downloading file:", requestedTorrent.Title, ":", newArray[0].Path())
            var index = -1
            for i, item := range requestedTorrent.Files() {
                if item.Path() == folderPath {
                    index = i + 1
                    break
                }
            }
            names := strings.Split(newArray[0].Path(), "/")
//             torr.Preload(requestedTorrent, index)
            c.Header("Content-Disposition", `attachment; filename="`+names[len(names) - 1]+`"`)
            c.Header("Content-Type", "application/octet-stream")
            requestedTorrent.Stream(index, c.Request, c.Writer)
        } else {
            c.Header("Content-Type", "text/html; charset=utf-8")
            folders := []string{}
            files := []string{}

            for _, item := range newArray {
                p := strings.TrimPrefix(item.Path(), folderPath + "/")
                if strings.Contains(p, "/") {
                    folders = append(folders, strings.Split(p, "/")[0])
                } else {
                    files = append(files, p)
                }
            }
            folders = sf.Unique(folders)
            c.String(200, listDir(path, folders, files))
        }
    }

}
