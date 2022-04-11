phony: readme

readme: #generates the README.md file
	@goreadme -functions -badge-godoc -badge-github=ci.yaml -badge-goreportcard -credit=false -skip-sub-packages > README.md