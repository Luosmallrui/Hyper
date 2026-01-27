package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"
	"strconv"

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
	products := r.Group("/v1/product")
	products.POST("/upload-product", authorize, context.Wrap(p.CreateProduct)) // 上传商品
	products.GET("/list-products", context.Wrap(p.BatchGetProducts))           // 列出商品
	products.GET("/product-details", context.Wrap(p.GetProductDetails))        // 获取商品详情
	//products.POST("delete-product/:productID", authorize, p.DeleteProduct) // 删除商品
	//products.GET("/", authorize, p.ListProducts)                           // 列出商品
}

func (p *ProductHandler) CreateProduct(c *gin.Context) error {
	var req types.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}

	// 1. 身份验证逻辑
	var partyId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		partyId = 3 // 调试模式固定为商家 3
	} else {
		// 💡 实际生产环境：从 JWT 中间件设置的 context 中获取 userID
		// userID, _ := c.Get("userID")
		// 然后根据 userID 查出其管理的 partyId
		partyId = int(req.PartyId) // 暂时从请求体取
	}

	if partyId == 0 {
		return response.NewError(http.StatusBadRequest, "未指定有效的商家ID")
	}

	// 2. 调用 Service
	err := p.ProductService.CreateProduct(c.Request.Context(), partyId, &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, req)
	return nil
}

func (p *ProductHandler) BatchGetProducts(c *gin.Context) error {
	partyId, _ := strconv.ParseInt(c.Query("party_id"), 10, 64)
	if partyId <= 0 {
		return response.NewError(http.StatusBadRequest, "请提供正确的商家ID")
	}

	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))

	// 调用 Service 获取数据
	res, err := p.ProductService.BatchGetProducts(c.Request.Context(), partyId, cursor, limit)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取列表失败")
	}

	response.Success(c, res)
	return nil
}

func (p *ProductHandler) GetProductDetails(c *gin.Context) error {
	idStr := c.Query("id")
	productID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || productID == 0 {
		return response.NewError(http.StatusBadRequest, "无效的商品ID")
	}
	paryIdstr := c.Query("party_id")
	partyId, _ := strconv.ParseInt(paryIdstr, 10, 64)
	if partyId <= 0 {
		return response.NewError(http.StatusBadRequest, "请提供正确的商家ID")
	}

	// 调用 Service 获取商品详情
	res, err := p.ProductService.GetDetailProduct(c.Request.Context(), uint64(productID), partyId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取商品详情失败")
	}

	response.Success(c, res)
	return nil
}
