package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Config         *config.Config
	ProductService service.IProductService
}

func NewProductHandler(config *config.Config, productService service.IProductService) *ProductHandler {
	return &ProductHandler{
		Config:         config,
		ProductService: productService,
	}
}
func (p *ProductHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(p.Config.Jwt.Secret))
	products := r.Group("/v1/products")
	products.POST("/upload-product", authorize, context.Wrap(p.CreateProduct)) // 上传商品
	//products.POST("delete-product/:productID", authorize, p.DeleteProduct) // 删除商品
	//products.GET("/", authorize, p.ListProducts)                           // 列出商品
}

func (p *ProductHandler) CreateProduct(c *gin.Context) error {
	var req types.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	err := p.ProductService.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, req)
	return nil

}
