# packer-plugin-s3

This plugin provides a simple provisioner, that retrieves objects from s3 and stores them at the given destination.

**Example**

```hcl
packer {
  required_plugins {
    s3 = {
      version = "2.0.7"
      source = "github.com/spacechunks/s3"
    }
  }
}

provisioner "s3" {
  profile = "<some-profile>" // optional
  objects {
    source = "mybucket/myfolder/somefile"
    destination = "/etc/myobject2"
  }
  objects {
    source = "mybucket2/myfolder/somefile"
    destination = "/etc/myobject2"
  }
}
```

**Configuration**

You can either define credentials in your AWS config file.
This example uses access and secret keys, but you can also define IAM credentials and SSO.

```
[profile test]
aws_access_key_id = my_secret_key
aws_secret_access_key = my_access_key
services = services
region = my_region

[services services]
s3 =
  endpoint_url = https://my_endpoint_url
```

or define them using environment variables

```
export AWS_ACCESS_KEY_ID=my_secret_key
export AWS_SECRET_ACCESS_KEY=my_access_key
export AWS_ENDPOINT_URL=https://my_endpoint_url
export AWS_REGION=my_region
```

## Development 

Running the tests:
* `make test` to run acceptance and unit tests
* `make test_unit` to only run unit tests
* `make test_acc` to only run acceptance tests (needs to have Docker installed)

Installing the plugin locally:
* `make install`

When changing the provisioners config you have to run `make gen` after, so the `hcl2spec.go` file gets generated.

If you encounter the following error while trying to build the plugin 

```
cannot use cty.Value{} (value of type cty.Value) as gob.GobEncoder value in variable declaration
```

checkout this link https://github.com/hashicorp/packer-plugin-sdk/issues/187
