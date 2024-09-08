# packer-plugin-s3

This plugin provides a simple provisioner, that retrieves objects from s3 and stores them at the given destination.

**Example**

```hcl
provisioner "s3" {
  access_key = "<my-access-key>"
  secret_key = "<my-secret-key>"
  endpoint = "ams1.vultrobjects.com" // do not include scheme i.e. https://
  source = "mybucket/myfolder/myobject"
  destination = "/etc/myobject"
  secure = false // defaults to true
}
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
