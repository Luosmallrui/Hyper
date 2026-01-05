package router

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/socket/handler"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.DebugMode)
}

// NewRouter 初始化配置路由
func NewRouter(conf *config.Config, handle *handler.Handler) *gin.Engine {

	router := gin.Default()
	router.Use(gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, err any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{"code": 500, "msg": "系统错误，请重试!!!"})
	}))

	authorize := middleware.Auth([]byte(conf.Jwt.Secret))

	router.GET("/wss", authorize, context.Wrap(handle.Chat.Conn))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": "success"})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, map[string]any{"msg": "请求地址不存在"})
	})

	debug := router.Group("/debug")
	{
		debug.GET("/", gin.WrapF(pprof.Index))
		debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/profile", gin.WrapF(pprof.Profile))
		debug.POST("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/trace", gin.WrapF(pprof.Trace))
		debug.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		debug.GET("/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		debug.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	return router
}
