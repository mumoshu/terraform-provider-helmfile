provider "helmfile" {}

resource "helmfile_embedding_example" "emb1" {
  embedded {
    path = "./helmfile.yaml"

    helm_binary = "helm3"



    working_directory = path.module

    environment = "default"

    environment_variables = {
      FOO = "emb1"
    }

    values = [
      <<EOF
{"name": "myapp"}
EOF
    ]

    selector = {
      labelkey1 = "value1"
    }

    kubeconfig = "kubeconfig"
  }
}

resource "helmfile_release_set" "mystack" {
  content = file("./helmfile.yaml")

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

  kubeconfig = "kubeconfig"
}

resource "helmfile_release_set" "mystack2" {
  content = <<EOF

releases:
- name: myapp2
  chart: sp/podinfo
  values:
  - image:
      tag: "123"
  labels:
    labelkey1: value1
- name: myapp3
  chart: sp/podinfo
  values:
  - image:
     tag: "2345"
EOF

  helm_binary = "helm3"

  //	working_directory = path.module
  working_directory = "mystack2"

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

  kubeconfig = "kubeconfig"
}

output "mystack_diff" {
  value = helmfile_release_set.mystack.diff_output
}

output "mystack_apply" {
  value = helmfile_release_set.mystack.apply_output
}

output "mystack2_diff" {
  value = helmfile_release_set.mystack2.diff_output
}

output "mystack2_apply" {
  value = helmfile_release_set.mystack2.apply_output
}

resource "helmfile_release" "myapp" {
  name = "myapp"
  namespace = "default"
  chart = "sp/podinfo"
  helm_binary = "helm3"

  //	working_directory = path.module
  //	working_directory = "myapp"
  values = [
    <<EOF
{ "image": {"tag": "3.1455" } }
EOF
  ]

  kubeconfig = "kubeconfig"
}


output "myapp_diff" {
  value = helmfile_release.myapp.diff_output
}

output "myapp_apply" {
  value = helmfile_release.myapp.apply_output
}

