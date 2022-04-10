package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// MyProvider .
type MyProvider struct{}

// Retrieve .
func (*MyProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}, nil
}

// IsExpired .
func (*MyProvider) IsExpired() bool { return false }

// CephMgmt .
type CephMgmt struct {
	BucketID  string
	blockSize uint32
	client    *s3.S3
}

// New .
func New(host, bucket string) *CephMgmt {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:           aws.String("default"),
			Endpoint:         aws.String(host),
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      credentials.NewCredentials(&MyProvider{}),
		},
	}))
	return &CephMgmt{
		BucketID:  bucket,
		blockSize: 8 * 1024 * 1024,
		client:    s3.New(sess),
	}
}

// Upload .
func (ceph *CephMgmt) Upload(srcPath, dstPath string) (*s3.PutObjectOutput, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	return ceph.client.PutObject(&s3.PutObjectInput{
		Body:   srcFile,
		Bucket: aws.String(ceph.BucketID),
		Key:    aws.String(dstPath),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
	})
}

// UploadMultiparts .
func (ceph *CephMgmt) UploadMultiparts(srcPath, dstPath string) (*s3.CompleteMultipartUploadOutput, error) {
	uploadInput := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(ceph.BucketID),
		Key:    aws.String(dstPath),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
	}
	uploadOutput, err := ceph.client.CreateMultipartUpload(uploadInput)
	if err != nil {
		return nil, err
	}

	// 分片上传
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()
	srcBufReader := bufio.NewReader(srcFile)

	var idx int64
	buf := make([]byte, ceph.blockSize)
	var partsCompleted []*s3.CompletedPart
	for idx = 1; ; idx++ {
		n, err := srcBufReader.Read(buf)
		if err == io.EOF {
			break
		}

		if uint32(n) != ceph.blockSize {
			buf = buf[:n]
		}
		uploadInput := &s3.UploadPartInput{
			Bucket:        aws.String(ceph.BucketID),
			Key:           aws.String(dstPath),
			PartNumber:    aws.Int64(idx),
			UploadId:      uploadOutput.UploadId,
			Body:          bytes.NewReader(buf),
			ContentLength: aws.Int64(int64(n)),
		}
		resp, err := ceph.client.UploadPart(uploadInput)
		if err != nil {
			return nil, err
		}

		partsCompleted = append(partsCompleted, &s3.CompletedPart{
			ETag:       resp.ETag,
			PartNumber: aws.Int64(idx),
		})
	}

	// 结束上传
	completeUploadInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(ceph.BucketID),
		Key:      aws.String(dstPath),
		UploadId: uploadOutput.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: partsCompleted,
		},
	}
	return ceph.client.CompleteMultipartUpload(completeUploadInput)
}

// Download .
func (ceph *CephMgmt) Download(srcPath, dstPath string) (int64, error) {
	resp, err := ceph.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(ceph.BucketID),
		Key:    aws.String(srcPath),
	})
	if err != nil {
		return 0, err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, resp.Body)
}

// Delete .
func (ceph *CephMgmt) Delete(key string) (*s3.DeleteObjectOutput, error) {
	return ceph.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(ceph.BucketID),
		Key:    aws.String(key),
	})
}

// List .
func (ceph *CephMgmt) List(prefix string) ([]string, error) {
	listInput := &s3.ListObjectsInput{
		Bucket: aws.String(ceph.BucketID),
	}
	if len(prefix) > 0 {
		listInput.Prefix = aws.String(prefix)
	}

	resp, err := ceph.client.ListObjects(listInput)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(resp.Contents))
	for _, content := range resp.Contents {
		key := aws.StringValue(content.Key)
		size := aws.Int64Value(content.Size)
		lines = append(lines, fmt.Sprintf("%s(%d)", key, size))
	}
	return lines, nil
}
