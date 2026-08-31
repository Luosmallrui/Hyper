package server

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/log"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type AppProvider struct {
	Config *config.Config
	Engine *gin.Engine
	DB     *gorm.DB
}

var (
	once sync.Once
	// 服务唯一ID
	serverId string
)

// InitServerId 使用配置端口初始化服务唯一ID，应在服务启动早期调用。
// 未调用时 GetServerId 会以默认端口 8083 延迟初始化（保持历史行为）。
func InitServerId(port int) {
	once.Do(func() {
		serverId = buildServerId(port)
	})
}

func GetServerId() string {
	once.Do(func() {
		serverId = buildServerId(8083)
	})
	return serverId
}

func buildServerId(port int) string {
	ip, err := getLocalIP() // 获取本机内网 IP
	if err != nil {
		log.L.Error("get local ip", zap.Error(err))
		return fmt.Sprintf("unknown:%d", port)
	}
	// 最终 sid 格式为: 192.168.1.10:8083
	return fmt.Sprintf("%s:%d", ip, port)
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// 检查 ip 网络地址，排除回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no ip address found")
}
func NewGinEngine(h *Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.Use(middleware.PrometheusMiddleware())
	r.Use(middleware.GinZap(), gin.Recovery())
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.HEAD("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// Serve the Mini Program's map web-view from the API host. Production
	// deploys its files to /root/map-h5; MAP_H5_DIR keeps the location explicit
	// for other environments without coupling the route to the process cwd.
	r.StaticFS("/map", http.Dir(mapH5Dir()))
	api := r.Group("/api")
	h.Auth.RegisterRouter(api)
	h.Map.RegisterRouter(api)
	h.User.RegisterRouter(api)
	h.Message.RegisterRouter(api)
	h.Note.RegisterRouter(api)
	h.Session.RegisterRouter(api)
	h.Follow.RegisterRouter(api)
	h.Group.RegisterRouter(api)
	h.GroupMember.RegisterRouter(api)
	h.CommentsHandler.RegisterRouter(api)
	h.TopicHandler.RegisterRouter(api)
	h.Pay.RegisterRouter(api)
	h.Party.RegisterRouter(api)
	h.ProductHandler.RegisterRouter(api)
	h.Points.RegisterRouter(api)
	h.Order.RegisterRouter(api)
	h.Serch.RegisterRouter(api)
	h.Channel.RegisterRouter(api)
	h.Event.RegisterRouter(api)
	h.Admin.RegisterRouter(api)
	h.Ticketing.RegisterRouter(api)
	h.Notification.RegisterRouter(api)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return r
}

func mapH5Dir() string {
	if dir := os.Getenv("MAP_H5_DIR"); dir != "" {
		return dir
	}
	if _, err := os.Stat("/root/map-h5/index.html"); err == nil {
		return "/root/map-h5"
	}
	return "./map-h5"
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置 CORS 头；回显请求 Origin，避免与 Allow-Credentials:true 冲突
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// 对于 OPTIONS 请求，直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Run(ctx *cli.Context, app *AppProvider) error {
	if !app.Config.Debug() {
		gin.SetMode(gin.ReleaseMode)
	}

	eg, groupCtx := errgroup.WithContext(ctx.Context)
	c := make(chan os.Signal, 1)
	// 终止的信号 服务要停止了
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	InitServerId(app.Config.Server.Http)

	log.L.Info("server starting", zap.String("serverId", serverId),
		zap.Int("port", app.Config.Server.Http),
		zap.String("env", "prod"),
	)
	if app.DB != nil {
		StartOrderCancelTask(app.DB, 15)
	}

	return run(c, eg, groupCtx, app)
}

func run(c chan os.Signal, eg *errgroup.Group, ctx context.Context, app *AppProvider) error {
	serv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Http),
		Handler: app.Engine,
	}

	// 启动 http 服务
	eg.Go(func() error {
		err := serv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		defer func() {
			log.L.Info("server stopping", zap.String("serverId", serverId))

			// 等待中断信号以优雅地关闭服务器
			timeCtx, timeCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer timeCancel()

			if err := serv.Shutdown(timeCtx); err != nil {
				log.L.Info("server stopping", zap.String("serverId", serverId), zap.Error(err))
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return nil
		}
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.L.Info("server stopping", zap.Error(err))
	}

	log.L.Info("server stopped", zap.String("serverId", serverId))

	return nil
}
