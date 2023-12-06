install-tools:
	brew install golangci-lint
	brew install caarlos0/tap/svu

release-patch:
	git tag -a $(shell svu patch)
	git push --tags

release-pypi:
	git tag -a $(shell svu patch --prefix 'testutils/python/')
	git push --tags

release-cargo:
	git tag -a $(shell svu patch --prefix 'testutils/rust/')
	git push --tags
