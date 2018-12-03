CMDS := $(GOPATH)/bin/houndd $(GOPATH)/bin/hound

SRCS := $(shell find . -type f -name '*.go')

WEBPACK_ARGS := -p
ifdef DEBUG
	WEBPACK_ARGS := -d
endif

ALL: $(CMDS)

ui: ui/bindata.go

node_modules:
	npm install

$(GOPATH)/bin/houndd: ui/bindata.go $(SRCS)
	go install github.com/gitgrep-com/gitgrep/cmds/houndd

$(GOPATH)/bin/hound: ui/bindata.go $(SRCS)
	go install github.com/gitgrep-com/gitgrep/cmds/hound

.build/bin/go-bindata:
	GOPATH=`pwd`/.build go get github.com/jteeuwen/go-bindata/...

ui/bindata.go: .build/bin/go-bindata node_modules $(wildcard ui/assets/**/*)
	rsync -r ui/assets/* .build/ui
	npx webpack $(WEBPACK_ARGS)
	$< -o $@ -pkg ui -prefix .build/ui -nomemcopy .build/ui/...

dev: ALL
	npm install

test:
	go test github.com/gitgrep-com/gitgrep/...

deploy-demo: ALL
	ssh demo "mv /opt/gitgrep/bin/houndd /opt/gitgrep/bin/houndd-old || true"
	scp $(GOPATH)/bin/houndd demo:/opt/gitgrep/bin/houndd
	ssh demo "sudo service gitgrep restart"

clean:
	rm -rf .build node_modules
