.PHONY: install
install: build
	cp build/bin/docker-workspace ${GOBIN}/docker-workspace

.PHONY: build
build: clean
	go build -o build/bin/docker-workspace cmd/docker-workspace-main.go

.PHONY: clean
clean:
	rm -rf build
