TEST?=./...

default: test

bin:
	@sh -c "$(CURDIR)/scripts/build.sh"

test:
	"$(CURDIR)/scripts/test.sh"

testrace:
	go test -race $(TEST) $(TESTARGS)


updatedeps:
	go get -d -v -p 2 ./...

run:
	go run main.go



.PHONY: bin default dev test updatedeps run
