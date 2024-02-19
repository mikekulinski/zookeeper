install:
	brew install golangci-lint
	git config core.hooksPath hooks
	chmod +x ./hooks/pre-commit
lint:
	golangci-lint run --fix