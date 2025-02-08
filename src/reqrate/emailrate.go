package reqrate

import (
	"context"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/emersion/go-imap/v2"
	"time"
)

func CheckIMAPRate(envelope *imap.Envelope, exp time.Duration) int64 {
	addressList := make([]imap.Address, 0, len(envelope.Sender)+len(envelope.From)+len(envelope.ReplyTo))
	addressList = append(addressList, envelope.Sender...)
	addressList = append(addressList, envelope.From...)
	addressList = append(addressList, envelope.ReplyTo...)

	return checkMailAddressListRate(addressList, exp)
}

func checkMailAddressListRate(addressList []imap.Address, exp time.Duration) int64 {
	var count int64 = -1
	var addressMap = make(map[string]bool, len(addressList))

	for _, address := range addressList {
		if yes, ok := addressMap[address.Addr()]; ok && yes {
			continue
		}

		res := checkMailAddressRate(address, exp)
		if res == -1 {
			continue
		} else if res > count {
			count = res
		}

		addressMap[address.Addr()] = true
	}

	return count
}

func CheckMailAddressListRate(addressList []*imap.Address, exp time.Duration) int64 {
	var count int64 = -1
	var addressMap = make(map[string]bool, len(addressList))

	for _, address := range addressList {
		if yes, ok := addressMap[address.Addr()]; ok && yes {
			continue
		}

		res := CheckMailAddressRate(address, exp)
		if res == -1 {
			continue
		} else if res > count {
			count = res
		}

		addressMap[address.Addr()] = true
	}

	return count
}

func checkMailAddressRate(address imap.Address, exp time.Duration) int64 {
	if address.Addr() == "" || !utils.IsValidEmail(address.Addr()) {
		return -1
	}

	key := fmt.Sprintf("req:email:[%s]", address.Addr())
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

func CheckMailAddressRate(address *imap.Address, exp time.Duration) int64 {
	if address.Addr() == "" || !utils.IsValidEmail(address.Addr()) {
		return -1
	}

	key := fmt.Sprintf("req:email:[%s]", address.Addr())
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

func CheckMailStringAddressRate(address string, exp time.Duration) int64 {
	if address == "" || !utils.IsValidEmail(address) {
		return -1
	}

	key := fmt.Sprintf("req:email:[%s]", address)
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
