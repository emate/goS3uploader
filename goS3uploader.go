// Author: Marcin Matlaszek https://github.com/emate
package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/exp/inotify"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"net/http"
	"os"
	"path"
)

func SendFile(filename string, AWSAuth aws.Auth, bucketName string, storePath string) {
	region := aws.USEast
	connection := s3.New(AWSAuth, region)
	bucket := connection.Bucket(bucketName)
	s3path := path.Join(storePath, path.Base(filename))
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	bytes := make([]byte, fileSize)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)

	filetype := http.DetectContentType(bytes)

	multi, err := bucket.InitMulti(s3path, filetype, s3.ACL("private"))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	const fileChunk = 52428800

	parts, err := multi.PutAll(file, fileChunk)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = multi.Complete(parts)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Done sending ", filename)
}

func S3Sender(events <-chan string, AWSAuth aws.Auth, bucketName string, storePath string) {
	for e := range events {
		fmt.Println("Sending ", e)
		go SendFile(e, AWSAuth, bucketName, storePath)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage %s -directory <DIRECTORY_TO_WATCH> -bucket-name <S3_BUCKET_NAME> -store-path <S3_STORE_PATH> ", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {

	flag.Usage = usage
	var directory = flag.String("directory", "", "Local directory to watch for files")
	var bucketName = flag.String("bucket-name", "", "S3 Bucket Name to store files")
	var storePath = flag.String("store-path", "", "S3 store path")

	flag.Parse()
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	events := make(chan string)

	AWSAuth := aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_SECRET_KEY"),
	}

	go func() {
		for {
			select {
			case event := <-watcher.Event:
				if event.Mask&inotify.IN_CLOSE_WRITE == inotify.IN_CLOSE_WRITE {
					events <- event.Name
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()
	go S3Sender(events, AWSAuth, *bucketName, *storePath)
	err = watcher.Watch(*directory)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
