packer {
  required_plugins {
    docker = {
      version = ">= 1.0.0"
      source = "github.com/hashicorp/docker"
    }
  }
}

source "docker" "test" {
  image = "alpine"
  commit = "true"
  run_command = ["-d", "-i", "-t", "{{.Image}}"]
}

variable "access_key" {
  default = env("S3_ACC_TEST_ACCESS_KEY")
}

variable "secret_key" {
  default = env("S3_ACC_TEST_SECRET_KEY")
}

variable "endpoint" {
  default = env("S3_ACC_TEST_ENDPOINT")
}

build {
  sources = [
    "source.docker.test"
  ]

  provisioner "s3" {
    access_key = var.access_key
    secret_key = var.secret_key
    endpoint = var.endpoint
    source = "s3-acc-test/file"
    destination = "/tmp/file"
  }
}