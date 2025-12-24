package service

import (
	"Hyper/config"
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"io"
	"time"
)

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

func NewOssService(cfg *config.OssConfig) IOssService {
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
	}
}

type OssService struct {
	Client     *oss.Client
	BucketName string
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
