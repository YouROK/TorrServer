package auth

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"server/log"
	"server/settings"
)

func SetupAuth(engine *gin.Engine) *gin.RouterGroup {
	if !settings.HttpAuth {
		return nil
	}
	accs := getAccounts()
	if accs == nil {
		return nil
	}
	return engine.Group("/", gin.BasicAuth(accs))
}

func getAccounts() gin.Accounts {
	buf, err := ioutil.ReadFile(filepath.Join(settings.Path, "accs.db"))
	if err != nil {
		return nil
	}
	var accs gin.Accounts
	err = json.Unmarshal(buf, &accs)
	if err != nil {
		log.TLogln("Error parse accs.db", err)
	}
	return accs
}
