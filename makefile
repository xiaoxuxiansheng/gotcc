.PHONY: cover
cover:
	go test -v -coverprofile="codecov.report" -covermode=atomic