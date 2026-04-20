.PHONY: install-wasm install-sig test

install-wasm:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ti.wasm .
	cp ti.wasm ../picopico-web-editor/worker
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js ../picopico-web-editor/worker

install-sig:
	bash ./shell/install_sigs.sh

test:
	bash ./shell/test.sh

%:
	@:

