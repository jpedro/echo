.PHONY: build
build:
	go build -o echo

.PHONY: test
test: build
	./echo --env dev

.PHONY: docker
docker:
	env GOOS=linux GOARCH=amd64 go build -o echo-linux-amd64
	docker build . -t echo

.PHONY: push
push: docker
	docker tag echo jpedrob/echo
	docker push jpedrob/echo
