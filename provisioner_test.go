package main

import (
	"testing"
)

func TestSourceFormat(t *testing.T) {
	tests := []struct {
		name      string
		conf      map[string]interface{}
		shouldErr bool
	}{
		{
			name: "valid source single object",
			conf: map[string]interface{}{
				"source":      "bucket/obj",
				"access_key":  "key",
				"secret_key":  "key",
				"endpoint":    "ams1.vultrobjects.com",
				"destination": "/dest",
			},
		},
		{
			name: "valid source nested object",
			conf: map[string]interface{}{
				"source":      "bucket/folder/obj",
				"access_key":  "key",
				"secret_key":  "key",
				"endpoint":    "ams1.vultrobjects.com",
				"destination": "/dest",
			},
		},
		{
			name: "missing bucket",
			conf: map[string]interface{}{
				"source":      "/obj",
				"access_key":  "key",
				"secret_key":  "key",
				"endpoint":    "ams1.vultrobjects.com",
				"destination": "/dest",
			},
			shouldErr: true,
		},
		{
			name: "no slashes",
			conf: map[string]interface{}{
				"source":      "obj",
				"access_key":  "key",
				"secret_key":  "key",
				"endpoint":    "ams1.vultrobjects.com",
				"destination": "/dest",
			},
			shouldErr: true,
		},
		{
			name: "missing obj",
			conf: map[string]interface{}{
				"source":      "obj/",
				"access_key":  "key",
				"secret_key":  "key",
				"endpoint":    "ams1.vultrobjects.com",
				"destination": "/dest",
			},
			shouldErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				p   = &S3Provisioner{}
				err = p.Prepare(tt.conf)
			)

			if tt.shouldErr && err == nil {
				t.Fatal("expected error but got nil")
				return
			}

			if !tt.shouldErr && err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
		})
	}
}
