.PHONY: test
test:
	go test ./cmd/labcon/app/...
	go test .

cov:
	gocov test ./cmd/labcon/app/... | gocov report
	gocov test . | gocov report

mock:
	go run ./cmd/labcon_mockgen/main.go
