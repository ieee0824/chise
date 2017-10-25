package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ieee0824/getenv"
)

type config struct {
	Bucket  string `json:"bucket"`
	Profile string `json:"profile"`
}

func loadConfig() (*config, error) {
	var ret = new(config)
	f, err := os.Open(filepath.Clean(fmt.Sprintf("%s/.chise/config.json", getenv.String("HOME"))))
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(f).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}

type uploader struct {
	bucketName string
	uploader   *s3manager.Uploader
}

func newUploader(c *config) *uploader {
	ret := new(uploader)
	ret.bucketName = c.Bucket
	cfg := new(aws.Config)
	cfg.Credentials = credentials.NewSharedCredentials("", c.Profile)
	ret.uploader = s3manager.NewUploader(session.New(cfg))

	return ret
}

func (u *uploader) upload(f *os.File, key string) error {
	_, err := u.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(u.bucketName),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String("image/png"),
	})
	return err
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalln(err)
	}
	u := newUploader(cfg)

	key := fmt.Sprintf("%v.png", time.Now().UnixNano())
	fileName := fmt.Sprintf("/tmp/%v", key)

	cmd := exec.Command("screencapture", "-i", fileName)

	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}

	defer func() {
		os.Remove(fileName)
	}()
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	if err := u.upload(f, "/"+key); err != nil {
		log.Fatalln(err)
	}

	result := fmt.Sprintf("https://%s/%s", cfg.Bucket, key)
	fmt.Println(result)

	if err := clipboard.WriteAll(result); err != nil {
		log.Fatalln(err)
	}
}
