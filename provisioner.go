package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest mapstructure-to-hcl2 -type Config,Object

type Config struct {
	AccessKey string   `mapstructure:"access_key"`
	SecretKey string   `mapstructure:"secret_key"`
	Endpoint  string   `mapstructure:"endpoint"`
	Objects   []Object `mapstructure:"objects"`
	Secure    *bool    `mapstructure:"secure" required:"false"`

	ctx interpolate.Context
}

type Object struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
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

	for _, o := range p.conf.Objects {
		var (
			err   = fmt.Errorf("invalid source %s: should be in the format of <bucket>/path/to/object", o.Source)
			parts = strings.Split(o.Source, "/")
		)

		if len(parts) < 2 {
			return err
		}

		// handle the following cases
		// * /obj
		// * obj/
		if parts[0] == "" || parts[1] == "" {
			return err
		}
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

	for _, o := range p.conf.Objects {
		src, err := interpolate.Render(o.Source, &p.conf.ctx)
		if err != nil {
			return fmt.Errorf("error interpolating source: %v", err)
		}

		dest, err := interpolate.Render(o.Destination, &p.conf.ctx)
		if err != nil {
			return fmt.Errorf("error interpolating destination: %v", err)
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
	}

	return nil
}
