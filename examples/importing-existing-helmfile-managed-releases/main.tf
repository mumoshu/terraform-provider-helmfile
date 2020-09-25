provider "helmfile" {}

resource "helmfile_release_set" "myapps" {
  content = file("./helmfile.yaml")
}
