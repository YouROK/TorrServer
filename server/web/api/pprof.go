package api

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func mountPprof(authorized gin.IRouter) {
	authorized.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	authorized.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	authorized.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	authorized.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	authorized.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	authorized.GET("/debug/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
	authorized.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	authorized.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	authorized.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	authorized.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	authorized.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}
