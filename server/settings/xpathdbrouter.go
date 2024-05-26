package settings

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"server/log"

	"golang.org/x/exp/slices"
)

type XPathDBRouter struct {
	dbs      []TorrServerDB
	routes   []string
	route2db map[string]TorrServerDB
	dbNames  map[TorrServerDB]string
}

func NewXPathDBRouter() *XPathDBRouter {
	router := &XPathDBRouter{
		dbs:      []TorrServerDB{},
		dbNames:  map[TorrServerDB]string{},
		routes:   []string{},
		route2db: map[string]TorrServerDB{},
	}
	return router
}

func (v *XPathDBRouter) RegisterRoute(db TorrServerDB, xPath string) error {
	newRoute := v.xPathToRoute(xPath)

	if slices.Contains(v.routes, newRoute) {
		return fmt.Errorf("route \"%s\" already in routing table", newRoute)
	}

	// First DB becomes Default DB with default route
	if len(v.dbs) == 0 && len(newRoute) != 0 {
		v.RegisterRoute(db, "")
	}

	if !slices.Contains(v.dbs, db) {
		v.dbs = append(v.dbs, db)
		v.dbNames[db] = reflect.TypeOf(db).Elem().Name()
		v.log(fmt.Sprintf("Registered new DB \"%s\", total %d DBs registered", v.getDBName(db), len(v.dbs)))
	}

	v.route2db[newRoute] = db
	v.routes = append(v.routes, newRoute)

	// Sort routes by length descending.
	//   It is important later to help selecting
	//   most suitable route in getDBForXPath(xPath)
	sort.Slice(v.routes, func(iLeft, iRight int) bool {
		return len(v.routes[iLeft]) > len(v.routes[iRight])
	})
	v.log(fmt.Sprintf("Registered new route \"%s\" for DB \"%s\", total %d routes", newRoute, v.getDBName(db), len(v.routes)))
	return nil
}

func (v *XPathDBRouter) xPathToRoute(xPath string) string {
	return strings.ToLower(strings.TrimSpace(xPath))
}

func (v *XPathDBRouter) getDBForXPath(xPath string) TorrServerDB {
	if len(v.dbs) == 0 {
		return nil
	}
	lookup_route := v.xPathToRoute(xPath)
	var db TorrServerDB = nil
	// Expected v.routes sorted by length descending
	for _, route_prefix := range v.routes {
		if strings.HasPrefix(lookup_route, route_prefix) {
			db = v.route2db[route_prefix]
			break
		}
	}
	return db
}

func (v *XPathDBRouter) Get(xPath, name string) []byte {
	return v.getDBForXPath(xPath).Get(xPath, name)
}

func (v *XPathDBRouter) Set(xPath, name string, value []byte) {
	v.getDBForXPath(xPath).Set(xPath, name, value)
}

func (v *XPathDBRouter) List(xPath string) []string {
	return v.getDBForXPath(xPath).List(xPath)
}

func (v *XPathDBRouter) Rem(xPath, name string) {
	v.getDBForXPath(xPath).Rem(xPath, name)
}

func (v *XPathDBRouter) CloseDB() {
	for _, db := range v.dbs {
		db.CloseDB()
	}
	v.dbs = nil
	v.routes = nil
	v.route2db = nil
	v.dbNames = nil
}

func (v *XPathDBRouter) getDBName(db TorrServerDB) string {
	return v.dbNames[db]
}

func (v *XPathDBRouter) log(s string, params ...interface{}) {
	if len(params) > 0 {
		log.TLogln(fmt.Sprintf("XPathDBRouter: %s: %s", s, fmt.Sprint(params...)))
	} else {
		log.TLogln(fmt.Sprintf("XPathDBRouter: %s", s))
	}
}
