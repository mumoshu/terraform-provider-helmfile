## Importing existing Helmfile project into Terraform

Let's say you have an existing releases that are managed by helmfile and `helmfile.yaml`:

```
$ helmfile -f helmfile.yaml apply 
```

You can migrate the releases into your terraform project by using `terraform import`.

First, you edit your .tf file to add a `helmfile_release_set`:

```hcl-terraform
resource "helmfile_release_set" "myapps" {
  content = file("./helmfile.yaml")
}
```

Run `terraform import` with the path to `helmfile.yaml` as the last argument:

```
$ terraform import helmfile_release_set.myapps ./helmfile.yaml
```

Run `terraform plan`: 

```
$ terraform plan

helmfile_release_set.myapps: Refreshing state... [id=btmjojkllhcl73m9no0g]

------------------------------------------------------------------------

No changes. Infrastructure is up-to-date.

This means that Terraform did not detect any differences between your
configuration and real physical resources that exist. As a result, no
actions need to be performed.
```

Ensure that there's no diff shown in the plan result.

If there's any, you should retry updating `resource "helmfile_releaset_set" "myapps"` in your .tf file and rerunning `terraform plan` until there is no diff anymore.
