
.PHONY: init
init:
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

.PHONY: fmt
# 格式化代码
fmt:
	gofumpt -w -l .
	goimports -w -l .


.PHONY: lint
# 格式化代码 lint
lint: fmt
	golangci-lint run -v -c .golangci.yaml --new-from-rev=HEAD~1


.PHONY: changelog
# 变更日志
changelog:
	git-chglog -o ./CHANGELOG.md


# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help