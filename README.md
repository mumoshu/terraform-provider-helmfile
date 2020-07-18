# terraform-provider-helmfile

Deploy [Helmfile](https://github.com/roboll/helmfile/) releases from within Terraform.

Benefits:

- Entry-point to your infrastructure and app deployments as a whole    
- Input terraform variables and outputs into Helmfile
- Blue-green deployment of your whole stack with tf's `create_before_destroy`

## Prerequisites

Install the `terraform-provider-helmfile` binary under `terraform.d/plugins/${OS}_${ARCH}`.

## Examples

There is nothing to configure for the provider, so you firstly declare the provider like:

```
provider "helmfile" {}
```

You can define a release in one of the three ways:

- Inline `helmfile_release`
- External `helmfile_release_set`
- Inline `helmfile_release_set`

`helmfile_release` would be a natural choice for users who are familiar with Terraform. It just map each Terraform `helmfile_release` resource to a Helm release 1-by-1:

```hcl
resource "helmfile_release" "myapp" {
	# `name` is the optional release name. When omitted, it's set to the ID of the resource, "myapp".
	# name = "myapp-${var.somevar}"
	namespace = "default"
	chart = "sp/podinfo"
	helm_binary = "helm3"

	working_directory = path.module
	values = [
		<<EOF
{ "image": {"tag": "3.14" } }
EOF
	]
}
```

External `helmfile_release_set` is the easiest way for existing Helmfile users, as the tf resource maps to the exsiting helmfile.yaml 1:1.

```
resource "helmfile_release_set" "mystack" {
    content = file("./helmfile.yaml")
}
```

The inline variant of the release set allows you to render helmfile.yaml without Go template but with the Terraform syntax:

```
resource "helmfile_release_set" "mystack" {
    # Install and choose from one of installed versions of helm
    # By changing this, you can upgrade helm per release_set
    # Default: helm
    helm_binary = "helm-3.0.0"

    # Install and choose from one of installed versions of helmfile
    # By changing this, you can upgrade helmfile per release_set
    # Default: helmfile
    binary = "helmfile-v0.93.0"

    working_directory = path.module

    # Maximum number of concurrent helm processes to run, 0 is unlimited (0 is a default value)
    concurrency = 0

    # Helmfile environment name to deploy
    # Default: default
    environment = "prod"

    # Environment variables available to helmfile's requireEnv and commands being run by helmfile
    environment_variables = {
        FOO = "foo"
        KUBECONFIG = "path/to/your/kubeconfig"
    }
    
    # State values to be passed to Helmfile
    values = {
      # Corresponds to --state-values-set name=myapp
      name = "myapp"
    }
    
    # State values files to be passed to Helmfile
    values = [
      file("overrides.yaml"),
      file("another.yaml"),
    ]
    
    # Label key-value pairs to filter releases 
    selector = {
      # Corresponds to -l labelkey1=value1
      labelkey1 = "value1"
    }
}

output "mystack_diff" {
  value = helmfile_release_set.mystack.diff_output
}

output "mystack_apply" {
  value = helmfile_release_set.mystack.apply_output
}
```

In the example above I am changing my working_directory, setting some environment variables that will be utilized by all my helmfiles.

Stdout and stderr from Helmfile runs are available in the debug log files. 

Running `terraform plan` runs `helmfile diff`.

It shows no changes if `helmfile diff` did not detect any changes:

```console
helmfile_release_set.mystack: Refreshing state... [id=bnd30hkllhcvvgsrplo0]

------------------------------------------------------------------------

No changes. Infrastructure is up-to-date.

This means that Terraform did not detect any differences between your
configuration and real physical resources that exist. As a result, no
actions need to be performed.
```

`terraform plan` surfaces changes in the `diff_output` field if `helmfile diff` detected any changes:

```
helmfile_release_set.mystack: Refreshing state... [id=bnd30hkllhcvvgsrplo0]

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # helmfile_release_set.mystack will be updated in-place
  ~ resource "helmfile_release_set" "mystack" {
        binary                = "helmfile"
      - diff_output           = "Comparing release=myapp-foo, chart=sp/podinfo\n\x1b[33mdefault, myapp-foo-podinfo, Deployment (apps) has changed:\x1b[0m\n  # Source: podinfo/templates/deployment.yaml\n  apiVersion: apps/v1\n  kind: Deployment\n  metadata:\n    name: myapp-foo-podinfo\n    labels:\n      app: podinfo\n      chart: podinfo-3.1.4\n      release: myapp-foo\n      heritage: Helm\n  spec:\n    replicas: 1\n    strategy:\n      type: RollingUpdate\n      rollingUpdate:\n        maxUnavailable: 1\n    selector:\n      matchLabels:\n        app: podinfo\n        release: myapp-foo\n    template:\n      metadata:\n        labels:\n          app: podinfo\n          release: myapp-foo\n        annotations:\n          prometheus.io/scrape: \"true\"\n          prometheus.io/port: \"9898\"\n      spec:\n        terminationGracePeriodSeconds: 30\n        containers:\n          - name: podinfo\n\x1b[31m-           image: \"stefanprodan/podinfo:foobar2aa\"\x1b[0m\n\x1b[32m+           image: \"stefanprodan/podinfo:foobar2a\"\x1b[0m\n            imagePullPolicy: IfNotPresent\n            command:\n              - ./podinfo\n              - --port=9898\n              - --port-metrics=9797\n              - --grpc-port=9999\n              - --grpc-service-name=podinfo\n              - --level=info\n              - --random-delay=false\n              - --random-error=false\n            env:\n            - name: PODINFO_UI_COLOR\n              value: cyan\n            ports:\n              - name: http\n                containerPort: 9898\n                protocol: TCP\n              - name: http-metrics\n                containerPort: 9797\n                protocol: TCP\n              - name: grpc\n                containerPort: 9999\n                protocol: TCP\n            livenessProbe:\n              exec:\n                command:\n                - podcli\n                - check\n                - http\n                - localhost:9898/healthz\n              initialDelaySeconds: 1\n              timeoutSeconds: 5\n            readinessProbe:\n              exec:\n                command:\n                - podcli\n                - check\n                - http\n                - localhost:9898/readyz\n              initialDelaySeconds: 1\n              timeoutSeconds: 5\n            volumeMounts:\n            - name: data\n              mountPath: /data\n            resources:\n              limits: null\n              requests:\n                cpu: 1m\n                memory: 16Mi\n        volumes:\n        - name: data\n          emptyDir: {}\n\nin ./helmfile.yaml: failed processing release myapp-foo: helm3 exited with status 2:\n  Error: identified at least one change, exiting with non-zero exit code (detailed-exitcode parameter enabled)\n  Error: plugin \"diff\" exited with error\n" -> null
      ~ dirty                 = true -> false
        environment           = "default"
        environment_variables = {
            "FOO" = "foo"
        }
        helm_binary           = "helm3"
        id                    = "bnd30hkllhcvvgsrplo0"
        path                  = "./helmfile.yaml"
        selector              = {
            "labelkey1" = "value1"
        }
        values                = {
            "name" = "myapp"
        }
        working_directory     = "."
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

Running `terraform apply` runs `helmfile apply` to deploy your releases.

The computed field `apply_output` is used to surface the output from Helmfile. You can use in the string interpolation to produce a useful Terraform output.

In the example below, the output `mystack_apply` is generated from `apply_output` so that you can review what has actually changed on `helmfile apply`: 

```console
helmfile_release_set.mystack: Refreshing state... [id=bnd30hkllhcvvgsrplo0]

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

mystack_apply = Comparing release=myapp-foo, chart=sp/podinfo
********************

	Release was not present in Helm.  Diff will show entire contents as new.

********************
...

mystack_diff = 
```

`terraform apply` just succeeds without any effect when there's no change detected by `helmfile`:

```console
helmfile_release_set.mystack: Refreshing state... [id=bnd30hkllhcvvgsrplo0]

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

mystack_apply =
mystack_diff =
```

## Develop
If you wish to build this yourself, follow the instructions:

	cd terraform-provider-helmfile
	go build

## Acknowledgement

The implementation of this product is highly inspired from [terraform-provider-shell](https://github.com/scottwinkler/terraform-provider-shell). A lot of thanks to the author!
