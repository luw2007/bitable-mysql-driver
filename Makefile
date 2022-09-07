
test:
	gotestsum --format=short-verbose -- -p=2 -coverprofile=cov.log ./driver
