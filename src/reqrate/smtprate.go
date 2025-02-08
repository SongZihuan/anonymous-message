package reqrate

import (
	"context"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"net/mail"
	"time"
)

func CheckSMTPSendAddressRate(sendType string, address *mail.Address, exp time.Duration) int64 {
	if sendType == "" || address.Address == "" || !utils.IsValidEmail(address.Address) {
		return -1
	}

	key := fmt.Sprintf("smtp:email:%s[%s]", sendType, address.Address)
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

func CheckSMTPSendAddressStringRate(sendType string, address string, exp time.Duration) int64 {
	if sendType == "" || address == "" || !utils.IsValidEmail(address) {
		return -1
	}

	key := fmt.Sprintf("smtp:email:%s[%s]", sendType, address)
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
