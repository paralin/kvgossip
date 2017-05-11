gengo: protogen install-go

buildarm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -v -a -o kvgossip ./cmd/kvgossip

docker:
	CGO_ENABLED=0 go build -v -o kvgossip ./cmd/kvgossip
	docker build -t "fuserobotics/kvgossip:latest" .

push: docker
	docker tag fuserobotics/kvgossip:latest registry.fusebot.io/fuserobotics/kvgossip:base
	docker push registry.fusebot.io/fuserobotics/kvgossip:base

pushl: docker
	docker tag fuserobotics/kvgossip:latest registry.fusebot.io/fuserobotics/kvgossip:latest
	docker push registry.fusebot.io/fuserobotics/kvgossip:latest

protogen:
	protowrap -I $${GOPATH}/src \
		--go_out=plugins=grpc:$${GOPATH}/src \
		--proto_path $${GOPATH}/src \
		--print_structure \
		--only_specified_files \
		$$(pwd)/**/*.proto

install-go:
	for D in */; do go install -v github.com/fuserobotics/kvgossip/$$D 2>/dev/null || true ; done
