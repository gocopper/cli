.PHONY: release
release:
	git tag -a $(version) -m "Release $(version)"
	git push origin $(version)
	goreleaser release --rm-dist


.PHONY: screenshots
screenshots:
	rm -rf out && mkdir out
	go run github.com/gocopper/cli/scripts/genscreenshots -pkg=cmd/copper -out=out
