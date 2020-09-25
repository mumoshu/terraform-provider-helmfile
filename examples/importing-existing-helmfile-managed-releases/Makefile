TF ?= terraform-v0.12.29
TEST_OUT_DIR ?= testout
TF_LOG ?= TRACE

.PHONY: dev
dev:
	rm tf.log tf.*.log || true
	rm -rf ./tf.*.pprof || true
	rm -rf terraform.tfstate
	make init import plan apply destroy

.PHONY: init
init:
	cd ../.. ; make build
	mkdir -p .terraform/plugins/darwin_amd64
	mv ../../dist/darwin_amd64/terraform-provider-helmfile .terraform/plugins/darwin_amd64/terraform-provider-helmfile
	$(TF) init

.PHONY: testout
testout:
	mkdir -p $(TEST_OUT_DIR)

.PHONY: apply
apply: testout
	TF_LOG=$(TF_LOG) TF_LOG_PATH=$(TEST_OUT_DIR)/apply.log time -l $(TF) apply -auto-approve

.PHONY: destroy
destroy: testout
	TF_LOG=$(TF_LOG) TF_LOG_PATH=$(TEST_OUT_DIR)/destroy.log time -l $(TF) destroy -auto-approve

.PHONY: plan
plan: testout
	TF_LOG=$(TF_LOG) TF_LOG_PATH=$(TEST_OUT_DIR)/plan.log time -l $(TF) plan

.PHONY: import
import: testout
	TF_LOG=$(TF_LOG) TF_LOG_PATH=$(TEST_OUT_DIR)/import.log time -l $(TF) import helmfile_release_set.myapps helmfile.yaml
