gengo: protogen install-go

protogen:
	protowrap -I $${GOPATH}/src \
		--gogo_out=plugins=grpc:$${GOPATH}/src \
		--proto_path $${GOPATH}/src \
		--print_structure \
		--only_specified_files \
		$$(pwd)/**/*.proto

deps:
	go get -u github.com/square/goprotowrap/cmd/protowrap

compile-roles:
	go build -v $(COMPILE_ROLES_DIR)

install-go:
	for D in */; do go install -v github.com/fuserobotics/kvgossip/$$D 2>/dev/null || true ; done
