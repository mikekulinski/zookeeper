PROTO_DIR := proto

install:
	# Install all dependencies using homebrew.
	brew install golangci-lint protobuf
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	# Set up our git hooks.
	git config core.hooksPath hooks
	chmod +x ./hooks/pre-commit

# Run the linter for our Go code.
lint:
	golangci-lint run --fix

# Regen all our protos.
generate:
	protoc \
		--go_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) \
		--go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) \
		$(shell find $(PROTO_DIR) -name '*.proto')

clean:
	rm -f $(PROTO_DIR)/*.pb.go