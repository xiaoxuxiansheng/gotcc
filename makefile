.PHONY: cover
cover:
	go test -v -race -coverprofile="codecov.report" -covermode=atomic