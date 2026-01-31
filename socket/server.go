package socket

import (
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"Hyper/socket/process"

	"Hyper/pkg/socket"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Hyper/config"
	"Hyper/socket/handler"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

var ErrServerClosed = errors.New("shutting down server")

type AppProvider struct {
	Config    *config.Config
	Engine    *gin.Engine
	Coroutine *process.Server
	Handler   *handler.Handler
	Db        *gorm.DB
	Redis     *redis.Client
	//Providers *client.Providers
}

func Run(ctx *cli.Context, app *AppProvider) error {
	eg, groupCtx := errgroup.WithContext(ctx.Context)

	if !app.Config.Debug() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 IM 渠道配置
	socket.Initialize(groupCtx, eg, func(name string) {
		//emailClient := app.Providers.EmailClient
		//if app.WechatPayConfig.App.Env == "prod" {
		//	_ = emailClient.SendMail(&email.Option{
		//		To:      app.WechatPayConfig.App.AdminEmail,
		//		Subject: fmt.Sprintf("[%s]守护进程异常", app.WechatPayConfig.App.Env),
		//		Body:    fmt.Sprintf("守护进程异常[%s]", name),
		//	})
		//}
	})

	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// 延时启动守护协程
	time.AfterFunc(3*time.Second, func() {
		app.Coroutine.Start(eg, groupCtx)
	})
	log.L.Info("server_id", zap.String("server_id", server.GetServerId()))
	log.L.Info("server Pid", zap.Any("server_pid", os.Getpid()))
	log.L.Info("server Version", zap.Any("Websocket Listen Port ", app.Config.Server.Websocket))

	return start(c, eg, groupCtx, app)
}

func start(c chan os.Signal, eg *errgroup.Group, ctx context.Context, app *AppProvider) error {
	serv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Websocket),
		Handler: app.Engine,
	}

	// 启动 Websocket 服务
	eg.Go(func() error {
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	eg.Go(func() (err error) {
		defer func() {
			log.L.Info("Shutting down component...")

			// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
			timeCtx, timeCancel := context.WithTimeout(context.TODO(), 3*time.Second)
			defer timeCancel()

			if err := serv.Shutdown(timeCtx); err != nil {
				log.L.Error("Server Shutdown Failed", zap.Error(err))
			}

			err = ErrServerClosed
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return nil
		}
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, ErrServerClosed) {
		log.L.Error("Server forced to shutdown", zap.Error(err))
	}

	log.L.Info("Server exiting")

	return nil
}
