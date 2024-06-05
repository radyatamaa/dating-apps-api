package helper

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/beego/i18n"
)

var ErrInvalidFormatJpeg = errors.New("format must be JPEG image")

func ItemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)
	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

// GetLangVersion sets site language version.
func GetLangVersion(ctx *beegoContext.Context) string {
	// 1. Check URL arguments.
	lang := ctx.Input.Query("lang")

	// Check again in case someone modifies on purpose.
	if !i18n.IsExist(lang) {
		lang = ""
	}

	// 2. Get language information from 'Accept-Language'.
	if len(lang) == 0 {
		al := ctx.Request.Header.Get("Accept-Language")
		if i18n.IsExist(al) {
			lang = al
		}
	}

	// 3. Default language is english.
	if len(lang) == 0 {
		lang = "en"
	}

	// Set language properties.
	return lang
}

func ValidateFile(fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 512) // 512 bytes should be enough to detect the file type
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	fileType := http.DetectContentType(buffer)
	if fileType != "image/jpeg" {
		return ErrInvalidFormatJpeg
	}

	return nil
}

func UploadFileJpeg(beegoCtx *beegoContext.Context,file multipart.File) (string,error)  {
	nameOfFile := fmt.Sprintf("%s.jpeg",GenerateRandomString(10))
	outputPath := fmt.Sprintf("external/storage")
	// Create the uploads folder if it doesn't exist
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return "",err
	}

	// Create the file on the server
	dst, err := os.Create(filepath.Join(outputPath, nameOfFile))
	if err != nil {
		return "",err
	}
	defer dst.Close()

	// Copy the uploaded file to the server
	_, err = io.Copy(dst, file)
	if err != nil {
		return "",err
	}
	outputPath = fmt.Sprintf("%s/%s",outputPath,nameOfFile)
	outputPath = fmt.Sprintf("%s://%s/%s" ,GetHttpOrHttps(beegoCtx), beegoCtx.Request.Host, outputPath)

	return outputPath,nil
}

func GetHttpOrHttps(beegoCtx *beegoContext.Context) string {
	// Get the original scheme from X-Forwarded-Proto header if available
	https := beegoCtx.Input.Header("X-Forwarded-Proto")
	if https == "" {
		// Fall back to the scheme from the request URL
		https = beegoCtx.Input.Scheme()
	}

	return https
}