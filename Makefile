ENV   ?= dev
STACK  = dilbert-feed-$(ENV)
FUNCS := $(subst /,,$(dir $(wildcard */main.go)))
RUST_FUNCS := $(subst /,,$(dir $(wildcard */lambda.rs)))
CDK   ?= ./node_modules/.bin/cdk

#
# deploy & destroy
#

dev: ENV=dev
dev: deploy

prod: ENV=prod
prod: deploy

deploy diff synth: build transpile
	@$(CDK) $@ $(STACK)

deploy: test

destroy: build transpile
	@$(CDK) destroy --force $(STACK)

bootstrap: build transpile
	@$(CDK) bootstrap

#
# build & transpile
#

build_funcs := $(FUNCS:%=build-%)

$(build_funcs):
	mkdir -p bin/$(@:build-%=%)
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=-buildid= -o bin/$(@:build-%=%)/handler ./$(@:build-%=%)

rust_funcs := $(RUST_FUNCS:%=rust-%)

rust: $(rust_funcs)

$(rust_funcs):
	cross build --release --target x86_64-unknown-linux-musl --bin $(@:rust-%=%)
	mkdir -p bin/$(@:rust-%=%)
	cp -f target/x86_64-unknown-linux-musl/release/$(@:rust-%=%) bin/$(@:rust-%=%)/bootstrap

build: $(build_funcs) $(rust_funcs)

transpile: node_modules
	@npm run build

node_modules:
	npm install

#
# lint
#

lint:
	go vet ./...
	golint -set_exit_status $$(go list ./...)

lint_funcs := $(FUNCS:%=lint-%)

$(lint_funcs):
	go vet ./$(@:lint-%=%)
	golint -set_exit_status ./$(@:lint-%=%)

#
# test
#

test:
	go test -v -cover -count=1 ./...

test_funcs := $(FUNCS:%=test-%)

$(test_funcs):
	go test -v -cover ./$(@:test-%=%)
