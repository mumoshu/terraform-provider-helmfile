package helmfile

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccHelmfileReleaseSet_basic(t *testing.T) {
	resourceName := "helmfile_release_set.the_product"
	releaseID := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHelmfileReleaseSetConfig_basic(releaseID),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "selector.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "diff_output", wantedHelmfileDiffOutputForReleaseID(releaseID)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccHelmfileReleaseSet_binaries(t *testing.T) {
	resourceName := "helmfile_release_set.the_product"
	releaseID := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHelmfileReleaseSetConfig_binaries(releaseID),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "selector.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "diff_output", wantedHelmfileDiffOutputForReleaseID(releaseID)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckShellScriptDestroy(s *terraform.State) error {
	_ = testAccProvider.Meta().(*ProviderInstance)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "helmfile_release_set" {
			continue
		}
		//
		//helmfileYaml := fmt.Sprintf("helmfile-%s.yaml", rs.Primary.ID)
		//
		//cmd := exec.Command("helmfile", "-f", helmfileYaml, "status")
		//if out, err := cmd.CombinedOutput(); err == nil {
		//	return fmt.Errorf("verifying helmfile status: releases still exist for %s", helmfileYaml)
		//} else if !strings.Contains(string(out), "Error: release: not found") {
		//	return fmt.Errorf("verifying helmfile status: unexpected error: %v:\n\nCOMBINED OUTPUT:\n%s", err, string(out))
		//}
	}
	return nil
}

func testAccHelmfileReleaseSetConfig_basic(randVal string) string {
	return fmt.Sprintf(`
resource "helmfile_release_set" "the_product" {
  content = <<EOF
repositories:
- name: sp
  url: https://stefanprodan.github.io/podinfo

releases:
- name: pi-%s
  chart: sp/podinfo
  values:
  - image:
      tag: "123"
  labels:
    labelkey1: value1
EOF

  helm_binary = "helm"

  kubeconfig = pathexpand("~/.kube/config")

  working_directory = "%s"

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
`, randVal, randVal)
}

func testAccHelmfileReleaseSetConfig_binaries(randVal string) string {
	return fmt.Sprintf(`
resource "helmfile_release_set" "the_product" {
  content = <<EOF
repositories:
- name: sp
  url: https://stefanprodan.github.io/podinfo

releases:
- name: pi-%s
  chart: sp/podinfo
  values:
  - image:
      tag: "123"
  labels:
    labelkey1: value1
EOF

  version = "0.128.1"
  helm_version = "3.2.1"

  kubeconfig = pathexpand("~/.kube/config")

  working_directory = "%s"

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
`, randVal, randVal)
}

func wantedHelmfileDiffOutputForReleaseID(id string) string {
	releaseName := fmt.Sprintf("pi-%s", id)

	return strings.ReplaceAll(`Adding repo sp https://stefanprodan.github.io/podinfo
"sp" has been added to your repositories

Comparing release=${RELEASE_NAME}, chart=sp/podinfo
********************

	Release was not present in Helm.  Diff will show entire contents as new.

********************
default, ${RELEASE_NAME}-podinfo, Deployment (apps) has been added:
- 
+ # Source: podinfo/templates/deployment.yaml
+ apiVersion: apps/v1
+ kind: Deployment
+ metadata:
+   name: ${RELEASE_NAME}-podinfo
+   labels:
+     app: ${RELEASE_NAME}-podinfo
+     chart: podinfo-4.0.6
+     release: ${RELEASE_NAME}
+     heritage: Helm
+ spec:
+   replicas: 1
+   strategy:
+     type: RollingUpdate
+     rollingUpdate:
+       maxUnavailable: 1
+   selector:
+     matchLabels:
+       app: ${RELEASE_NAME}-podinfo
+   template:
+     metadata:
+       labels:
+         app: ${RELEASE_NAME}-podinfo
+       annotations:
+         prometheus.io/scrape: "true"
+         prometheus.io/port: "9898"
+     spec:
+       terminationGracePeriodSeconds: 30
+       containers:
+         - name: podinfo
+           image: "stefanprodan/podinfo:123"
+           imagePullPolicy: IfNotPresent
+           command:
+             - ./podinfo
+             - --port=9898
+             - --port-metrics=9797
+             - --grpc-port=9999
+             - --grpc-service-name=podinfo
+             - --level=info
+             - --random-delay=false
+             - --random-error=false
+           env:
+           - name: PODINFO_UI_COLOR
+             value: #34577c
+           ports:
+             - name: http
+               containerPort: 9898
+               protocol: TCP
+             - name: http-metrics
+               containerPort: 9797
+               protocol: TCP
+             - name: grpc
+               containerPort: 9999
+               protocol: TCP
+           livenessProbe:
+             exec:
+               command:
+               - podcli
+               - check
+               - http
+               - localhost:9898/healthz
+             initialDelaySeconds: 1
+             timeoutSeconds: 5
+           readinessProbe:
+             exec:
+               command:
+               - podcli
+               - check
+               - http
+               - localhost:9898/readyz
+             initialDelaySeconds: 1
+             timeoutSeconds: 5
+           volumeMounts:
+           - name: data
+             mountPath: /data
+           resources:
+             limits: null
+             requests:
+               cpu: 1m
+               memory: 16Mi
+       volumes:
+       - name: data
+         emptyDir: {}
default, ${RELEASE_NAME}-podinfo, Service (v1) has been added:
- 
+ # Source: podinfo/templates/service.yaml
+ apiVersion: v1
+ kind: Service
+ metadata:
+   name: ${RELEASE_NAME}-podinfo
+   labels:
+     app: podinfo
+     chart: podinfo-4.0.6
+     release: ${RELEASE_NAME}
+     heritage: Helm
+ spec:
+   type: ClusterIP
+   ports:
+     - port: 9898
+       targetPort: http
+       protocol: TCP
+       name: http
+     - port: 9999
+       targetPort: grpc
+       protocol: TCP
+       name: grpc
+   selector:
+     app: ${RELEASE_NAME}-podinfo

Affected releases are:
  ${RELEASE_NAME} (sp/podinfo) UPDATED

Identified at least one change
`, "${RELEASE_NAME}", releaseName)
}
