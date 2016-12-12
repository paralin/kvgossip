gengo: protogen install-go

dumb-init:
	wget https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 -O dumb-init

docker: dumb-init
	CGO_ENABLED=0 go build -v -o kvgossip ./cmd/kvgossip
	docker build -t "fuserobotics/kvgossip:latest" .

protogen:
	protowrap -I $${GOPATH}/src \
		--go_out=plugins=grpc:$${GOPATH}/src \
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
