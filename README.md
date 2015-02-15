# gos3uploader
Simple program in go to watch for inotify events and put files on S3 immediately after close (IN_CLOSE_WRITE inotify flag).

### Requirements

Due to usage of inotify which is Linux-specific, goS3uploader runs only on Linux. It depends on couple of external packages:

```
	golang.org/x/exp/inotify
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
```

### Installation

```sh
$ git clone  https://github.com/emate/goS3uploader
$ cd goS3uploader && go build goS3uploader.go
```

### How to run

```sh
$ export AWS_ACCESS_KEY="<YOUR_ACCESS_KEY>"
$ export AWS_SECRET_KEY="<YOUR_SECRET_KEY>"
$ goS3uploader -directory <LOCAL_DIRECTORY_TO_OBSERVE> -bucket-name <S3_BUCKET_NAME> -store-path <S3_STORE_PATH> 
```
