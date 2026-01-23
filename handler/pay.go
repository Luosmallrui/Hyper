package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/service"
	"github.com/gin-gonic/gin"
)

type Pay struct {
	Config     *config.Config
	PayService service.IPayService
}

func (p *Pay) RegisterRouter(r gin.IRouter) {
	//authorize := middleware.Auth([]byte(f.Config.Jwt.Secret))
	pay := r.Group("/v1/pay")
	pay.POST("/", context.Wrap(p.RepairPay))
}

func (p *Pay) RepairPay(c *gin.Context) error {
	return nil
}
