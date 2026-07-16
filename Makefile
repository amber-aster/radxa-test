APP := radxa-dashboard-test
GO ?= go

.PHONY: tidy build build-radxa run clean

tidy:
	$(GO) mod tidy

build: tidy
	$(GO) build -trimpath -ldflags "-s -w" -o $(APP) .

build-radxa: tidy
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 $(GO) build -trimpath -ldflags "-s -w" -o $(APP)-linux-arm64 .

run: build
	./$(APP) -fullscreen=false

clean:
	rm -f $(APP) $(APP)-linux-arm64
