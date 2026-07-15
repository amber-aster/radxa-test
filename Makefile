APP := radxa-dashboard-test
GO ?= go

.PHONY: tidy build run clean

tidy:
	$(GO) mod tidy

build: tidy
	$(GO) build -trimpath -ldflags "-s -w" -o $(APP) .

run: build
	./$(APP) -fullscreen=true -stress=true

clean:
	rm -f $(APP)
