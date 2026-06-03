package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultS3DownloadPartSize = int64(8 * 1024 * 1024)

func (s *S3Storage) Restore(key string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	headOutput, err := s.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &bucketKey,
	})
	if err != nil {
		_ = tempFile.Close()
		return nil, err
	}

	// Download in 8 MiB ranges to match aws-cli style segmented GET requests.
	if headOutput.ContentLength > 0 {
		for start := int64(0); start < headOutput.ContentLength; start += defaultS3DownloadPartSize {
			end := start + defaultS3DownloadPartSize - 1
			if end >= headOutput.ContentLength {
				end = headOutput.ContentLength - 1
			}

			byteRange := fmt.Sprintf("bytes=%d-%d", start, end)
			input := &s3.GetObjectInput{
				Bucket: &s.Bucket,
				Key:    &bucketKey,
				Range:  &byteRange,
			}
			if headOutput.ETag != nil {
				input.IfMatch = headOutput.ETag
			}

			output, err := s.Client.GetObject(context.TODO(), input)
			if err != nil {
				_ = tempFile.Close()
				return nil, err
			}

			_, err = io.Copy(tempFile, output.Body)
			closeErr := output.Body.Close()
			if err != nil {
				_ = tempFile.Close()
				return nil, err
			}

			if closeErr != nil {
				_ = tempFile.Close()
				return nil, closeErr
			}
		}
	} else {
		output, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &s.Bucket,
			Key:    &bucketKey,
		})
		if err != nil {
			_ = tempFile.Close()
			return nil, err
		}

		_, err = io.Copy(tempFile, output.Body)
		closeErr := output.Body.Close()
		if err != nil {
			_ = tempFile.Close()
			return nil, err
		}

		if closeErr != nil {
			_ = tempFile.Close()
			return nil, closeErr
		}
	}

	return tempFile, tempFile.Close()
}
