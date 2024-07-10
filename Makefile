.PHONY: snapshot
.PHONY: build
.PHONY: release
.PHONY: patch
.PHONY: run
.PHONY: clean
.PHONY: tidy

snapshot: tidy
	goreleaser build --clean --snapshot

build: tidy
	goreleaser build --clean

release: tidy
	git tag "$(shell svu next)"
	git push --tags
	goreleaser release --clean

patch: tidy
	git tag "$(shell svu next --force-patch-increment)"
	git push --tags
	goreleaser release --clean

tidy:
	go mod tidy

run: snapshot
	go run .

clean:
	go clean
