package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/freggy/packers3/testdata"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
	"os/exec"
	"testing"
)

func TestAccS3Basic(t *testing.T) {
	var (
		accessKey = envOrDie(t, "S3_ACC_TEST_ACCESS_KEY")
		secretKey = envOrDie(t, "S3_ACC_TEST_SECRET_KEY")
		endpoint  = envOrDie(t, "S3_ACC_TEST_ENDPOINT")
		ctx       = context.Background()
		bucket    = "s3-acc-test"
		objName   = "file"
	)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		t.Fatalf("failed to create s3 client: %v", err)
	}

	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatalf("failed to create s3 bucket: %v", err)
	}

	if _, err := client.PutObject(
		ctx,
		bucket,
		objName,
		bytes.NewReader([]byte(testdata.Template)),
		-1,
		minio.PutObjectOptions{},
	); err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	tc := &acctest.PluginTestCase{
		Name:     "s3_basic_test",
		Template: testdata.Template,
		Teardown: func() error {
			if err := client.RemoveObject(ctx, bucket, objName, minio.RemoveObjectOptions{}); err != nil {
				return fmt.Errorf("failed to remove object: %v", err)
			}
			if err := client.RemoveBucket(ctx, bucket); err != nil {
				return fmt.Errorf("failed to remove bucket: %v", err)
			}
			return nil
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code. logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, tc)
}

func envOrDie(t *testing.T, key string) string {
	env := os.Getenv(key)
	if env == "" {
		t.Fatalf("environment variable %s not set", key)
	}
	return env
}
