package server

import (
	"Hyper/config"
	"Hyper/pkg/response"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
		id, err := gonanoid.Generate("0123456789abcdefghjklmnpqrstuvwxyz", 10)
		if err != nil {
			panic(err)
		}

		serverId = id
	})
}

func NewGinEngine(h *Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(response.ErrorMiddleware())
	h.Auth.RegisterRouter(r)

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
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	log.Printf("HTTP Listen Port :%d", app.Config.Server.Http)
	log.Printf("HTTP Server Pid  :%d", os.Getpid())

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
			log.Println("Shutting down serv...")

			// 等待中断信号以优雅地关闭服务器
			timeCtx, timeCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer timeCancel()

			if err := serv.Shutdown(timeCtx); err != nil {
				log.Fatalf("HTTP Server Shutdown Err: %s", err)
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
		log.Fatalf("HTTP Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")

	return nil
}
