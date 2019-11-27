provider "helmfile" {}

resource "helmfile_release_set" "mystack" {
	path = "./helmfile.yaml"

	helm_binary = "helm3"

	working_directory = path.module

	environment = "default"

	environment_variables = {
		FOO = "foo"
	}

	values = [
		<<EOF
{"name": "myapp"}
EOF
	]

	selector = {
	  labelkey1 = "value1"
	}
}

resource "helmfile_release_set" "mystack2" {
	content = <<EOF
releases:
- name: myapp2
  chart: sp/podinfo
EOF

	helm_binary = "helm3"

	working_directory = path.module

	environment = "default"

	environment_variables = {
		FOO = "foo"
	}

	values = [
		<<EOF
{"name": "myapp"}
EOF
	]

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

resource "helmfile_release" "myapp" {
	name = "myapp"
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


output "myapp_diff" {
	value = helmfile_release.myapp.diff_output
}

output "myapp_apply" {
	value = helmfile_release.myapp.apply_output
}

