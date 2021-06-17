GOVARS = GOOS=linux GOARCH=amd64
TIME = $(shell date +'%Y-%m-%d-%H%M%S')
REPO = jpedrob
NAME = echo
TAG ?= $(TIME)
VERSION = $(shell cat version.txt)

.PHONY: test
test:
	go build -o $(NAME)
	./$(NAME)

.PHONY: image
image:
	$(GOVARS) go build -o $(NAME)
	docker build . -t $(NAME)
	docker tag $(NAME) $(REPO)/$(NAME)
	docker tag $(NAME) $(REPO)/$(NAME):$(TAG)
	docker push $(REPO)/$(NAME)
	docker push $(REPO)/$(NAME):$(TAG)
	$(shell echo $(TAG) > version.txt)

.PHONY: deploy
deploy: image
	kubectl apply -f k8s/app.yaml

.PHONY: update
update: image
	kubectl set image deployment/$(NAME) app=$(REPO)/$(NAME):$(VERSION)
