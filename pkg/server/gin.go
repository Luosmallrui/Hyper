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
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type AppProvider struct {
	Config *config.Config
	Engine *gin.Engine
}

var (
	once sync.Once
	// 服务唯一ID
	serverId string
)

func init() {
	once.Do(func() {
		ip, err := getLocalIP() // 获取本机内网 IP
		if err != nil {
			log.L.Fatal("get local ip", zap.Error(err))
		}
		// 最终 sid 格式为: 192.168.1.10:8083
		serverId = fmt.Sprintf("%s:%d", ip, 8083)
	})
}
func GetServerId() string {
	return serverId
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
	r.Use(middleware.GinZap(), gin.Recovery())
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
	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置 CORS 头
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有来源
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

	log.L.Info("server starting", zap.String("serverId", serverId),
		zap.Int("port", app.Config.Server.Http),
		zap.String("env", "prod"),
	)

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
