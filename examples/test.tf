provider "helmfile" {}

resource "helmfile_release_set" "mystack" {
	path = "./helmfile.yaml"

	helm_binary = "helm3"

	working_directory = path.module

	environment = "default"

	environment_variables = {
		FOO = "foo"
	}

	values = {
	  name = "myapp"
	}

	selector = {
	  labelkey1 = "value1"
	}
}

output "mystack_diff" {
  value = helmfile_release_set.mystack.diff_output
}

output "mystack_apply" {
  value = helmfile_release_set.mystack.apply_output
}
