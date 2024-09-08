package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strings"
)

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest mapstructure-to-hcl2 -type Config

var ErrInvalidSource = errors.New("invalid source: should be in the format of <bucket>/path/to/object")

type Config struct {
	AccessKey   string `mapstructure:"access_key"`
	SecretKey   string `mapstructure:"secret_key"`
	Endpoint    string `mapstructure:"endpoint"`
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
	Secure      *bool  `mapstructure:"secure" required:"false"`

	ctx interpolate.Context
}

type S3Provisioner struct {
	conf Config
}

func (p *S3Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.conf.FlatMapstructure().HCL2Spec()
}

func (p *S3Provisioner) Prepare(raws ...interface{}) error {
	if err := config.Decode(&p.conf, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.conf.ctx,
	}, raws...); err != nil {
		return err
	}

	parts := strings.Split(p.conf.Source, "/")
	if len(parts) < 2 {
		return ErrInvalidSource
	}

	// handle the following cases
	// * /obj
	// * obj/
	if parts[0] == "" || parts[1] == "" {
		return ErrInvalidSource
	}

	return nil
}

func (p *S3Provisioner) Provision(
	ctx context.Context,
	ui packer.Ui,
	communicator packer.Communicator,
	m map[string]interface{},
) error {
	p.conf.ctx.Data = m

	accessKey, err := interpolate.Render(p.conf.AccessKey, &p.conf.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating access key: %v", err)
	}

	secretKey, err := interpolate.Render(p.conf.SecretKey, &p.conf.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating secret key: %v", err)
	}

	endpoint, err := interpolate.Render(p.conf.Endpoint, &p.conf.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating endpoint: %v", err)
	}

	src, err := interpolate.Render(p.conf.Source, &p.conf.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %v", err)
	}

	dest, err := interpolate.Render(p.conf.Destination, &p.conf.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating destination: %v", err)
	}

	secure := true
	if p.conf.Secure != nil {
		secure = *p.conf.Secure
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return fmt.Errorf("error creating s3 client: %v", err)
	}

	var (
		parts  = strings.Split(src, "/")
		bucket = parts[0]
		path   = strings.Join(parts[1:], "/")
	)

	ui.Sayf("retrieving object %s from bucket %s", path, bucket)

	obj, err := client.GetObject(ctx, bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("error retrieving object %s: %v", path, err)
	}

	if err := communicator.Upload(dest, obj, nil); err != nil {
		return fmt.Errorf("error uploading object %s: %v", path, err)
	}

	return nil
}
