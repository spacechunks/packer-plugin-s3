package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/acctest"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/freggy/packers3/testdata"
)

func TestAccS3Basic(t *testing.T) {
	var (
		ctx     = context.Background()
		bucket  = "s3-acc-test"
		objName = "dir/file1"
		content = "this-is-file-content"
	)

	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		func(opts *awsconfig.LoadOptions) error {
			opts.SharedConfigProfile = "test"
			opts.SharedConfigFiles = []string{"./testdata/__aws_config"}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	tests := []*acctest.PluginTestCase{
		{
			Name:     "s3_profile_basic_test",
			Init:     true,
			Template: testdata.ProfileTemplate,
			BuildExtraArgs: []string{
				fmt.Sprintf("-var=expected_content=%s", content),
			},
			Setup: func() error {
				if err := setupBucket(ctx, client, bucket, objName, content); err != nil {
					return err
				}
				return os.Setenv("AWS_CONFIG_FILE", "testdata/__aws_config")
			},
			Teardown: func() error {
				return teardown(ctx, client, bucket, objName)
			},
			Check: check,
		},
		{
			Name:     "s3_env_basic_test",
			Template: testdata.EnvTemplate,
			Init:     true,
			BuildExtraArgs: []string{
				fmt.Sprintf("-var=expected_content=%s", content),
			},
			Setup: func() error {
				if err := setupBucket(ctx, client, bucket, objName, content); err != nil {
					return err
				}
				if err = os.Setenv("AWS_ACCESS_KEY_ID", envOrDie(t, "S3_ACC_TEST_ACCESS_KEY")); err != nil {
					return err
				}
				if err := os.Setenv("AWS_SECRET_ACCESS_KEY", envOrDie(t, "S3_ACC_TEST_SECRET_KEY")); err != nil {
					return err
				}
				if err := os.Setenv("AWS_ENDPOINT_URL", envOrDie(t, "S3_ACC_TEST_ENDPOINT")); err != nil {
					return err
				}
				if err := os.Setenv("AWS_REGION", envOrDie(t, "S3_ACC_TEST_REGION")); err != nil {
					return err
				}
				return nil
			},
			Teardown: func() error {
				return teardown(ctx, client, bucket, objName)
			},
			Check: check,
		},
	}

	for _, tc := range tests {
		acctest.TestPlugin(t, tc)
	}
}

func setupBucket(ctx context.Context, client *s3.Client, bucket, obj, content string) error {
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return fmt.Errorf("failed to create s3 bucket: %v", err)
	}

	if _, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &obj,
		Body:   bytes.NewReader([]byte(content)),
	}); err != nil {
		return fmt.Errorf("failed to put object: %v", err)
	}
	return nil
}

func teardown(ctx context.Context, client *s3.Client, bucket, obj string) error {
	if _, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &obj,
	}); err != nil {
		return fmt.Errorf("failed to remove object: %v", err)
	}

	if _, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return fmt.Errorf("failed to delete bucket: %v", err)
	}

	// some providers (hetzner *cough* cough*) are a bit slow deleting
	// the bucket from their database. even though the remove call succeeds
	// creating will fail, due to BucketAlreadyExists. to work around this
	// we wait a little bit
	time.Sleep(1 * time.Second)
	return nil
}

func check(buildCommand *exec.Cmd, logfile string) error {
	if buildCommand.ProcessState != nil {
		if buildCommand.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("bad exit code. logfile: %s", logfile)
		}
	}
	return nil
}

func envOrDie(t *testing.T, key string) string {
	env := os.Getenv(key)
	if env == "" {
		t.Fatalf("environment variable %s not set", key)
	}
	return env
}
