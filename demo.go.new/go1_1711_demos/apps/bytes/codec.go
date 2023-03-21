package bytes

import (
	"encoding/binary"
	"go1_1711_demo/utils"
)

//
// bytes encoding and decoding: 读写固定格式的 blob 数据到 bytes queue 中。
//
// blob: timestamp:8byte + hash:8byte + len(key):2byte + key + data
//

const (
	timestampSize = 8
	hashSize      = 8
	keySize       = 2
	// header: timestamp + hash + key
	HeaderSize = timestampSize + hashSize + keySize
)

func WrapBlob(timestamp, hash uint64, key string, data []byte) []byte {
	keyLength := len(key)
	blobLen := HeaderSize + keyLength + len(data)

	blob := make([]byte, blobLen)
	binary.LittleEndian.PutUint64(blob, timestamp)
	binary.LittleEndian.PutUint64(blob[timestampSize:], hash)
	binary.LittleEndian.PutUint16(blob[timestampSize+hashSize:], uint16(keyLength))

	copy(blob[HeaderSize:], key)
	copy(blob[HeaderSize+keyLength:], data)
	return blob
}

func ReadTimestampFromBlob(blob []byte) uint64 {
	return binary.LittleEndian.Uint64(blob)
}

func ReadHashFromBlob(blob []byte) uint64 {
	return binary.LittleEndian.Uint64(blob[timestampSize:])
}

func ReadKeyFromBlob(blob []byte) string {
	length := binary.LittleEndian.Uint16(blob[timestampSize+hashSize:])
	dst := make([]byte, length)
	copy(dst, blob[HeaderSize:HeaderSize+length])
	return utils.Bytes2str(dst)
}

func ReadDataFromBlob(blob []byte) []byte {
	length := binary.LittleEndian.Uint16(blob[timestampSize+hashSize:])
	dst := make([]byte, len(blob)-int(HeaderSize+length))
	copy(dst, blob[HeaderSize+length:])
	return dst
}

// GetIntBytesSize returns number of bytes for input int.
func GetIntBytesSize(num int) int {
	switch {
	case num < (1<<7 - 1): // 127
		return 1
	case num < (1<<14 - 2): // 16382
		return 2
	case num < (1<<21 - 3):
		return 3
	default:
		return 4
	}
}
