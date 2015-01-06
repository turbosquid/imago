package s3

import (
	"os"
	"path"
	"testing"
)

var key = os.Getenv("AWS_KEY")
var secret = os.Getenv("AWS_SECRET")
var download_uri = os.Getenv("DOWNLOAD_URI")
var upload_uri = os.Getenv("UPLOAD_URI")
var local_file = path.Base(download_uri)

func TestConnect(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	if !conn.IsConnected() {
		t.Errorf("Could not connect to S3 -- we will need ListMyBuckets permission for this to work")
	}
}

func TestDownload(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	err := conn.DownloadFile(download_uri, local_file)
	if err != nil {
		t.Errorf("Unable to download file: %s (%s)", local_file, err.Error())
	}
}

func TestUpload(t *testing.T) {
	conn := New(key, secret, "us-east-1")
	err := conn.UploadFile(local_file, upload_uri, "application/txt")
	if err != nil {
		t.Errorf("Unable to upload file: %s", err.Error())
	}
}
