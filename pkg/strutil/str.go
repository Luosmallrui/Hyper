package strutil

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenValidateCode 生成数字验证码（使用 crypto/rand，避免可预测）
func GenValidateCode(length int) string {
	var sb strings.Builder
	max := big.NewInt(10)
	for i := 0; i < length; i++ {
		n, err := crand.Int(crand.Reader, max)
		if err != nil {
			// crypto/rand 失败时退化为时间戳尾数，保证不 panic
			sb.WriteByte(byte('0' + time.Now().UnixNano()%10))
			continue
		}
		sb.WriteByte(byte('0' + n.Int64()))
	}
	return sb.String()
}

// Random 生成随机字符串
func Random(length int) string {
	var result []byte
	bytes := []byte("0123456789abcdefghijklmnopqrstuvwxyz")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	return string(result)
}

// MtSubstr 字符串截取
func MtSubstr(value string, start, end int) string {

	if start > end {
		return ""
	}

	str := []rune(value)

	if length := len(str); end > length {
		end = length
	}

	return string(str[start:end])
}

func BoolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

// FileSuffix 获取文件后缀名
func FileSuffix(filename string) string {
	return strings.TrimPrefix(path.Ext(filename), ".")
}

func NewMsgId() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func NewUuid() string {
	return uuid.New().String()
}
