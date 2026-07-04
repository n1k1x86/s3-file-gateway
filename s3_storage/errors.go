package s3_storage

import "github.com/aws/aws-sdk-go-v2/service/s3/types"

var (
	ErrNoSuchBucket *types.NoSuchBucket
	ErrNoSuchKey    *types.NoSuchKey
)
