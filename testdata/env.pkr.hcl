packer {
  required_plugins {
    docker = {
      version = ">= 1.0.0"
      source = "github.com/hashicorp/docker"
    }
  }
}

variable "expected_content" {
  type = string
}

source "docker" "test" {
  image = "alpine"
  commit = "true"
  run_command = ["-d", "-i", "-t", "{{.Image}}"]
}

build {
  sources = [
    "source.docker.test"
  ]

  provisioner "s3" {
    objects {
      source      = "s3-acc-test/dir/file1"
      destination = "/tmp/file1"
    }
    objects {
      source      = "s3-acc-test/dir/file1"
      destination = "/tmp/file2"
    }
  }

  provisioner "shell" {
    inline = [
      "[[ $(cat /tmp/file1) == ${var.expected_content} ]] && exit 0 || exit 1",
      "[[ $(cat /tmp/file2) == ${var.expected_content} ]] && exit 0 || exit 1",
    ]
  }
}