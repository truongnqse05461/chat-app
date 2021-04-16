package aws

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	// "net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/hiyali/logli"
)

const (
	S3_REGION = "ap-southeast-1"
	S3_BUCKET = "chat-app-storage"
	S3_ACL    = "public-read"
)

type S3Handler struct {
	Session *session.Session
	Bucket  string
}

func UploadHandler(c *gin.Context)  {
	//var w http.ResponseWriter = c.Writer
	//var r *http.Request = c.Request
	//if len(os.Args) != 3 {
	//	log.FatalF("usage: %s <filename> <s3-filepath>\n", filepath.Base(os.Args[0]))
	//}

	//filename := os.Args[1]
	//key := os.Args[2]

	filename := "D:/ChatAppGolang-master/script.sql"
	key := "profile-image/script.sql"

	file, err := os.Open(filename)
	if err != nil {
		log.FatalF("os.Open - filename: %v, err: %v", filename, err)
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.FatalF("session.NewSession - filename: %v, err: %v", filename, err)
	}

	handler := S3Handler{
		Session: sess,
		Bucket:  S3_BUCKET,
	}

	// contents, err := handler.ReadFile(filename)
	// if err != nil {
	// 	log.FatalF("ReadFile - filename: %v, err: %v", filename, err)
	// }

	_, err = handler.ReadFile(key)

	//err = handler.UploadFile(key, filename)
	if err != nil {
		log.FatalF("UploadFile - filename: %v, err: %v", filename, err)
	}
	log.Info("UploadFile - success")
}

func main() {
	if len(os.Args) != 3 {
		log.FatalF("usage: %s <filename> <s3-filepath>\n", filepath.Base(os.Args[0]))
	}

	filename := os.Args[1]
	key := os.Args[2]

	file, err := os.Open(filename)
	if err != nil {
		log.FatalF("os.Open - filename: %v, err: %v", filename, err)
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.FatalF("session.NewSession - filename: %v, err: %v", filename, err)
	}

	handler := S3Handler{
		Session: sess,
		Bucket:  S3_BUCKET,
	}

	// contents, err := handler.ReadFile(filename)
	// if err != nil {
	// 	log.FatalF("ReadFile - filename: %v, err: %v", filename, err)
	// }

	err = handler.UploadFile(key, filename)

	if err != nil {
		log.FatalF("UploadFile - filename: %v, err: %v", filename, err)
	}
	log.Info("UploadFile - success")
}

func (h S3Handler) UploadFile(key string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.FatalF("os.Open - filename: %s, err: %v", filename, err)
	}
	defer file.Close()

	// buffer := []byte(body)

	_, err = s3.New(h.Session).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(h.Bucket),
		Key:                aws.String(key),
		ACL:                aws.String(S3_ACL),
		Body:               file, // bytes.NewReader(buffer),
		ContentDisposition: aws.String("attachment"),
		// ContentLength:      aws.Int64(int64(len(buffer))),
		// ContentType:        aws.String(http.DetectContentType(buffer)),
		// ServerSideEncryption: aws.String("AES256"),
	})

	// log.DebugF("s3.New - res: %v", res)
	return err
}

func (h S3Handler) ReadFile(key string) (string, error) {
	results, err := s3.New(h.Session).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(h.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
