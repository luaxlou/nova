package novaoss

import aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"

var _ func() (*aliyunoss.Bucket, error) = Bucket
