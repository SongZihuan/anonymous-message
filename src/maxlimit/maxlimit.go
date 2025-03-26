// Copyright 2025 AnonymousMessage Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package maxlimit

// MAX_BYTES_LIMIT 10 MB = 10485760 Bytes
const MAX_BYTES_LIMIT = 10485760

func StringTooBig(data string) bool {
	return DataTooBig([]byte(data))
}

func DataTooBig(data []byte) bool {
	return len(data) > MAX_BYTES_LIMIT
}
