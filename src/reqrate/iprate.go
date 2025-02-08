package reqrate

import (
	"context"
	"fmt"
	"time"
)

func CheckHttpReqIP(ip string, exp time.Duration) int64 {
	key := fmt.Sprintf("req:ip:[%s]", ip)
	res := rdb.Incr(context.Background(), key)
	if res.Err() != nil {
		return -1
	}

	count := res.Val()

	if count == 1 {
		if ok := rdb.Expire(context.Background(), key, exp).Val(); !ok {
			return -1
		}
	}

	return count
}
