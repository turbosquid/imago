package s3

import (
	"os"
	"testing"
)

var key = os.Getenv("AWS_KEY")
var secret = os.Getenv("AWS_SECRET")

func TestConnect(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	if !conn.IsConnected() {
		t.Errorf("Could not connect to S3")
	}
}

func TestUpload(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	err := conn.UploadFile("test.txt", "s3://dev-test-mikey.hero3d.net/go-test/test.txt", "application/txt")
	if err != nil {
		t.Errorf("Unable to upload file: %s", err.Error())
	}
}

func TestDownload(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	err := conn.DownloadFile("s3://dev-test-mikey.hero3d.net/go-test/test.txt", "test_downloaded.txt")
	if err != nil {
		t.Errorf("Unable to upload file: %s", err.Error())
	}
}
