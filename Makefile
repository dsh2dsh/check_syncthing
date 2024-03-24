PROGNAME=  check_syncthing
TEST_ARGS=

test:
	go test ${TEST_ARGS} ./...

test-e2e:
	go test ${TEST_ARGS} -tags=e2e ./...

build:
	go build -ldflags="-s -w" ./

clean:
	rm -f "${PROGNAME}"

${PROGNAME}: build

api/http.gen.go: api/http.gen.yaml api/syncthing.yaml
	oapi-codegen -config api/http.gen.yaml api/syncthing.yaml
