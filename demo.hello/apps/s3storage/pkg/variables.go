package pkg

import (
	"os"
)

var (
	host      string
	bucket    string
	accessKey string
	secretKey string
)

func init() {
	host = os.Getenv("S3_HOST")
	bucket = os.Getenv("S3_BUCKET")
	accessKey = os.Getenv("S3_ACCESS_KEY")
	secretKey = os.Getenv("S3_SECRET_KEY")

	if len(host) == 0 {
		panic("Env variables is not set")
	}
}
