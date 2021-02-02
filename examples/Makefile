TF ?= terraform-v0.12.29
TEST_OUT_DIR ?= testout/$(shell cat USE | tr -d '\n' || echo current)
TF_LOG ?= TRACE

.PHONY: dev
dev:
	rm tf.log tf.*.log || true
	rm -rf ./tf.*.pprof || true
	kind get kubeconfig > kubeconfig
	make testdata/01-bootstrap init plan apply
	make testdata/02-more-resources plan apply
	make testdata/03-add-prom plan apply
	make testdata/04-issue50 plan apply
	make destroy

.PHONY: init
init:
	cd .. ; make build
	mv ../dist/darwin_amd64/terraform-provider-helmfile .terraform/plugins/darwin_amd64/terraform-provider-helmfile
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

STEPS = $(wildcard testdata/*)

.PHONY: $(STEPS)
.SECONDEXPANSION:
$(STEPS):%:$$(wildcard %/*.yaml)
	@echo Preparing $@
	cp $@/*.{tf,yaml} .
	@echo $(shell basename $@) > USE
