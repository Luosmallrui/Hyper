package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"strings"

	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

type OssService struct {
	Client     *oss.Client
	BucketName string
	ImageRepo  *dao.Image
}

var _ IOssService = (*OssService)(nil)

type IOssService interface {
	// Upload 上传本地文件
	Upload(ctx context.Context, localPath, objectKey string) error

	// UploadReader 上传流（HTTP / 表单上传）
	UploadReader(ctx context.Context, reader io.Reader, objectKey string) error

	// Download 下载到本地
	Download(ctx context.Context, objectKey, localPath string) error

	// DownloadReader 下载为流
	DownloadReader(ctx context.Context, objectKey string) (io.ReadCloser, error)

	// Delete 删除对象
	Delete(ctx context.Context, objectKey string) error

	// SignURL 生成临时访问 URL（秒）
	SignURL(ctx context.Context, objectKey string, expireSeconds int64) (string, error)

	ListBuckets(ctx context.Context) ([]string, error)
	UploadImage(ctx context.Context, userID int, header *multipart.FileHeader) (*types.UploadImageResp, error)
}

func (s *OssService) UploadImage(ctx context.Context, userID int, header *multipart.FileHeader) (*types.UploadImageResp, error) {

	const maxSize int64 = 10 << 20 // 10MB

	if header == nil {
		return nil, fmt.Errorf("missing image")
	}
	// header.Size 不可信，但可做第一道拦截
	if header.Size <= 0 || header.Size > maxSize {
		return nil, fmt.Errorf("image size invalid")
	}

	f, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 要能 Seek，否则无法在“读头校验/取尺寸”后再上传同一份流
	seeker, ok := f.(io.ReadSeeker)
	if !ok {
		return nil, fmt.Errorf("uploaded file is not seekable")
	}

	// 1) MIME 校验（读取前 512 bytes）
	head := make([]byte, 512)
	n, _ := seeker.Read(head)
	contentType := http.DetectContentType(head[:n])
	allowedMime := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedMime[contentType] {
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}
	_, _ = seeker.Seek(0, io.SeekStart)

	// 2) 读取尺寸 + 格式（不解码全图）
	cfg, format, err := image.DecodeConfig(seeker)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}
	format = strings.ToLower(format)
	allowedFmt := map[string]bool{"jpeg": true, "png": true, "webp": true}
	if !allowedFmt[format] {
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
	_, _ = seeker.Seek(0, io.SeekStart)

	// 3) 生成 ID / objectKey
	imageID := snowflake.GenID()
	ext := "." + format
	if format == "jpeg" {
		ext = ".jpg"
	}
	objectKey := fmt.Sprintf("note/%s/%d%s",
		time.Now().Format("2006/01/02"),
		imageID,
		ext,
	)

	// 4) 上传 OSS（强制限制读取）
	limited := io.LimitReader(seeker, maxSize+1)

	// 只保留一种：PutObject（示例，按你 OSS SDK 调整）
	if _, err := s.Client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
		Body:   limited,
	}); err != nil {
		return nil, err
	}

	// 5) 写 image 表（status=uploaded）
	img := models.Image{
		ID:        imageID, // BIGINT
		UserID:    userID,
		OssKey:    objectKey,
		Width:     cfg.Width,
		Height:    cfg.Height,
		Status:    types.ImageStatusUploaded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.ImageRepo.CreateImage(ctx, &img)
	if err != nil {
		return nil, err
	}
	return &types.UploadImageResp{
		ImageID: imageID,
		Url:     "https://cdn.hypercn.cn/" + objectKey,
		Width:   cfg.Width,
		Height:  cfg.Height,
	}, nil
}

// ListBuckets 列举当前账号下所有 Bucket
func (s *OssService) ListBuckets(
	ctx context.Context,
) ([]string, error) {

	out, err := s.Client.ListBuckets(ctx, &oss.ListBucketsRequest{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		if b.Name != nil {
			buckets = append(buckets, *b.Name)
		}
	}

	return buckets, nil
}

func NewOssService(cfg *config.OssConfig, imageRepo *dao.Image) IOssService {
	ossCfg := oss.LoadDefaultConfig().
		WithEndpoint(cfg.Endpoint).
		WithRegion(cfg.Region).
		WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.AccessKeySecret,
			),
		)

	client := oss.NewClient(ossCfg)

	return &OssService{
		Client:     client,
		BucketName: cfg.Bucket,
		ImageRepo:  imageRepo,
	}
}

// Upload 上传本地文件
func (s *OssService) Upload(
	ctx context.Context,
	localPath, objectKey string,
) error {

	_, err := s.Client.PutObjectFromFile(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
	}, localPath)
	return err
}

// UploadReader 上传 Reader（HTTP 上传场景）
func (s *OssService) UploadReader(
	ctx context.Context,
	reader io.Reader,
	objectKey string,
) error {
	_, err := s.Client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
		Body:   reader,
	})
	return err
}

// Download 下载到本地文件
func (s *OssService) Download(
	ctx context.Context,
	objectKey, localPath string,
) error {

	_, err := s.Client.GetObjectToFile(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
	}, localPath)
	return err
}

// DownloadReader
func (s *OssService) DownloadReader(
	ctx context.Context,
	objectKey string,
) (io.ReadCloser, error) {

	out, err := s.Client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
	})
	if err != nil {
		return nil, err
	}

	return out.Body, nil
}

// Delete 删除对象
func (s *OssService) Delete(
	ctx context.Context,
	objectKey string,
) error {

	_, err := s.Client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
	})
	return err
}

// SignURL 生成临时访问 URL
func (s *OssService) SignURL(
	ctx context.Context,
	objectKey string,
	expireSeconds int64,
) (string, error) {

	result, err := s.Client.Presign(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.BucketName),
		Key:    oss.Ptr(objectKey),
	}, oss.PresignExpires(time.Duration(expireSeconds)))
	if err != nil {
		return "", err
	}

	return result.URL, nil
}
