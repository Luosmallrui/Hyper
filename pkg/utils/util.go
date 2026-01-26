package utils

import (
	"Hyper/pkg/context"
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/speps/go-hashids/v2"
)

// MtRand 生成指定范围内的随机数
func MtRand(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}

func PanicTrace(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}
func GenHashID(salt string, id int) string {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 12
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{id})
	return e
}

func GetQueryOrTokenUserID(c *gin.Context) (int, error) {
	if v := c.Query("user_id"); v != "" {
		return strconv.Atoi(v)
	}
	uid, err := context.GetUserID(c)
	return int(uid), err
}
