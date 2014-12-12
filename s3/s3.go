package s3

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"io"
	"io/ioutil"
	//	"log"
	"os"
	"regexp"
)

const BUF_LEN = 1 << 20 // 1 MB

var s3Regexp = regexp.MustCompile(`s3://(.+?)(/.+)`)

type S3Connection struct {
	connection *s3.S3
}

func New(key string, secret string, region string) *S3Connection {
	var conn S3Connection
	var auth aws.Auth
	auth.AccessKey = key
	auth.SecretKey = secret
	conn.connection = s3.New(auth, aws.GetRegion(region))
	return &conn
}

func (s *S3Connection) UploadFile(local string, remote string, contenttype string) (err error) {
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	bucketname, key := parseS3Uri(remote)
	bucket := s.connection.Bucket(bucketname)
	var opts s3.Options
	fi, err := f.Stat()
	err = bucket.PutReader(key, f, fi.Size(), contenttype, s3.BucketOwnerFull, opts)
	defer f.Close()
	return err
}

func fileExists(fn string) bool {
	if _, err := os.Stat(fn); err == nil {
		return true
	}
	return false
}

func getEtag(bucket *s3.Bucket, key string) (etag string, err error) {
	headers := make(map[string][]string)
	resp, err := bucket.Head(key, headers)
	if err != nil {
		return "", err
	}
	etag = resp.Header["Etag"][0]
	return etag, err
}

func checkFileExistence(bucket *s3.Bucket, key string, local string) bool {
	if fileExists(local) && fileExists(local+".etag") {
		remote_etag, err := getEtag(bucket, key)
		if err != nil {
			return false
		}
		local_etag, _ := ioutil.ReadFile(local + ".etag")
		if string(local_etag) == remote_etag {
			return true
		}
	}
	return false
}

func (s *S3Connection) DownloadFile(remote string, local string) (err error) {
	bucketname, key := parseS3Uri(remote)
	bucket := s.connection.Bucket(bucketname)
	if checkFileExistence(bucket, key, local) {
		return nil
	}

	f, err := os.Create(local)
	if err != nil {
		return err
	}
	defer f.Close()

	etag, err := getEtag(bucket, key)
	if err != nil {
		return err
	}
	reader, err := bucket.GetReader(key)
	if err != nil {
		return err
	}
	defer reader.Close()
	buf := make([]byte, BUF_LEN)
	for {
		n, ferr := reader.Read(buf)
		if ferr != nil && ferr != io.EOF {
			err = ferr
			break
		}
		if n == 0 {
			break
		}
		if _, err := f.Write(buf[:n]); err != nil {
			break
		}
	}
	err = ioutil.WriteFile(local+".etag", []byte(etag), 0644)
	return err
}

func (s *S3Connection) IsConnected() bool {
	_, err := s.connection.GetService()
	if err != nil {
		return false
	} else {
		return true
	}
}

func parseS3Uri(uri string) (bucket string, key string) {
	matches := s3Regexp.FindStringSubmatch(uri)
	if matches != nil {
		bucket = matches[1]
		key = matches[2]
	}
	return bucket, key
}
