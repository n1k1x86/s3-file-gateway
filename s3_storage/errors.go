package s3_storage

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func IsErrNoSuchBucket(err error) bool {
	var errNoSuchBucket *types.NoSuchBucket
	return errors.As(err, &errNoSuchBucket)
}

func IsErrNoSuchKey(err error) bool {
	var errNoSuchKey *types.NoSuchKey
	return errors.As(err, &errNoSuchKey)
}
