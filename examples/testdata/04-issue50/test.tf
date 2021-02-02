provider "helmfile" {}

resource "helmfile_release_set" "issue50_scenario1" {
  working_directory = "../examples/issue50/scenario1/platform"
  binary = "helmfile-0.137.0"
  content = file("issue50/scenario1/platform/helmfile.yaml")
  kubeconfig = "kubeconfig"
  values = [
    <<-EOF
    case: issue50_scenario1
    EOF
  ]
}

resource "helmfile_release_set" "issue50_scenario2" {
  working_directory = "../examples/issue50/scenario2/platform"
  kubeconfig = "kubeconfig"
  values = [
    <<-EOF
    case: issue50_scenario2
    EOF
  ]
}
