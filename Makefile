phony: readme lint

readme: #generates the README.md file
	@goreadme -functions -badge-godoc -badge-github=ci.yaml -badge-goreportcard -credit=false -skip-sub-packages > README.md

lint: #runs the go linter
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
	@golangci-lint run