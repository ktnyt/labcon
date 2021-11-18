.PHONY: test
test:
	go test ./cmd/labcon/app/...

cov:
	gocov test ./cmd/labcon/app/... | gocov report

mock:
	go run ./cmd/labcon_mockgen/main.go
