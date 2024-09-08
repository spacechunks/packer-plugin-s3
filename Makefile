BIN := packer-plugin-s3

all: gen install
test: test_unit test_acc

.PHONY: gen
gen:
	go generate .

.PHONY: test_acc
test_acc:
	PACKER_ACC=1 go test . -run TestAcc

.PHONY: test_unit
test_unit:
	go test

install: $(BIN)
	packer plugins install --path ./packer-plugin-s3 "github.com/freggy/s3"
	rm $(BIN)

$(BIN):
	go build -o packer-plugin-s3