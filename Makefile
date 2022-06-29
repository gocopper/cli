.PHONY: release
release:
	git tag -a $(version) -m "Release $(version)"
	git push origin $(version)
	goreleaser release --rm-dist

