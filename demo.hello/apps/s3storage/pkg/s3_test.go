package pkg

import (
	"fmt"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	ceph := New(host, bucket)
	files, err := ceph.List("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("files count:", len(files))
	fmt.Println("files:", files)
}

func TestUpload(t *testing.T) {
	ceph := New(host, bucket)
	resp, err := ceph.Upload("/tmp/test/log.txt", "test/log.txt")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("upload success, etag:", resp.ETag)

	files, err := ceph.List("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("files count:", len(files))
	fmt.Println("files:", files)
}

func TestDownload(t *testing.T) {
	ceph := New(host, bucket)
	n, err := ceph.Download("test/storagefile.test", "/tmp/test/storagefile.test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("download success, file size: %dM", (n / 1024 / 1024))
}

func TestDelete(t *testing.T) {
	ceph := New(host, bucket)
	if _, err := ceph.Delete("test/storagefile.test"); err != nil {
		t.Fatal(err)
	}

	files, err := ceph.List("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("files count:", len(files))
	fmt.Println("files:", strings.Join(files, ","))
}

func TestUploadMultiparts(t *testing.T) {
	ceph := New(host, bucket)
	resp, err := ceph.UploadMultiparts("/tmp/test/storagefile.test", "test/storagefile.test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)

	files, err := ceph.List("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("files count:", len(files))
	fmt.Println("files:", files)
}
