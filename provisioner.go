package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest mapstructure-to-hcl2 -type Config,Object

type Config struct {
	Profile      string   `mapstructure:"profile"`
	UsePathStyle bool     `mapstructure:"use_path_style"`
	Objects      []Object `mapstructure:"objects"`

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

	var cfg aws.Config
	if p.conf.Profile != "" {
		tmp, err := awsconfig.LoadDefaultConfig(
			context.Background(),
			awsconfig.WithSharedConfigProfile(p.conf.Profile),
		)
		if err != nil {
			return fmt.Errorf("error loading aws config from profile: %v", err)
		}
		cfg = tmp
	} else {
		tmp, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("error loading aws config: %v", err)
		}
		cfg = tmp
	}

	downloader := manager.NewDownloader(s3.NewFromConfig(cfg))

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
			parts   = strings.Split(src, "/")
			bucket  = parts[0]
			path    = strings.Join(parts[1:], "/")
			writeAt = manager.NewWriteAtBuffer(make([]byte, 0))
		)

		ui.Sayf("retrieving object %s from bucket %s", path, bucket)

		if _, err := downloader.Download(ctx, writeAt, &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &path,
		}); err != nil {
			return fmt.Errorf("error downloading object %s: %v", path, err)
		}

		if err := communicator.Upload(dest, bytes.NewReader(writeAt.Bytes()), nil); err != nil {
			return fmt.Errorf("error uploading object %s: %v", path, err)
		}
	}

	return nil
}
