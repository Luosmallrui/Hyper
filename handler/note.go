package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path"
	"strings"
	"time"
)

type Note struct {
	OssService service.IOssService
	Config     *config.Config
}

func (n *Note) RegisterRouter(r gin.IRouter) {
	g := r.Group("/v1/note")
	g.POST("/upload", context.Wrap(n.UploadImage))
}
func (n *Note) UploadImage(c *gin.Context) error {
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(500, err.Error())
	}
	const maxSize = 10 << 20
	if header.Size > maxSize {
		return fmt.Errorf("image size exceeds 10MB")
	}

	file, err := header.Open()
	if err != nil {
		return response.NewError(500, err.Error())
	}
	defer file.Close()
	var (
		width  int
		height int
	)

	if seeker, ok := file.(io.Seeker); ok {
		cfg, format, err := image.DecodeConfig(file)
		if err == nil {
			width = cfg.Width
			height = cfg.Height
		}
		allowed := map[string]bool{
			"jpeg": true,
			"png":  true,
			"webp": true,
		}
		if !allowed[strings.ToLower(format)] {
			return fmt.Errorf("unsupported image format")
		}
		seeker.Seek(0, 0) // 重置指针
	}

	objectKey := fmt.Sprintf("note/%s/%s%s", time.Now().Format("2006/01/02"), uuid.NewString(),
		path.Ext(header.Filename),
	)

	if err := n.OssService.UploadReader(c.Request.Context(), file, objectKey); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, types.UploadResponse{
		Key: fmt.Sprintf("https://%s.%s/%s",
			n.Config.Oss.Bucket, n.Config.Oss.Endpoint,
			objectKey),
		Width:  width,
		Height: height,
	})
	return nil
}
