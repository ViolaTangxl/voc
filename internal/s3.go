package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"viola/voc/utils"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	sess *session.Session
	svc  *s3.S3
)

func s3Init() {

	sess, _ = session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(utils.ACCESS_KEY, utils.SECRET_KEY, ""),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false), //virtual-host style方式，不要修改
	})

	svc = s3.New(sess)
}

// ListBuckets 列出获取所有桶
func ListBuckets() {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Fatalf("报错啦：%v\n", err)
	}

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	for _, b := range result.Buckets {
		fmt.Printf("%s\n", aws.StringValue(b.Name))
	}
}

// UploadFile 上传文件
func uploadFile(bucket, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Unable to open file %q, %v", err)
	}

	defer file.Close()

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		log.Fatalf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func s3Start() {
	s3Init()
	taobaoFolder := "../taobao/"
	files, err := ioutil.ReadDir(taobaoFolder)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			csvFiles := filepath.Join(taobaoFolder, file.Name())
			uploadFile("violatang", csvFiles)
		}
	}

}
