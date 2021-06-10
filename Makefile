SHA1 = $(shell git --no-pager log -1 --format=%h)
TIME = $(shell date +'%Y-%m-%d-%H%M%S')
NAME = echo
TAG? = $(TIME) #-$(SHA1)
REPO = jpedrob
### Stupid make re-evaluates $(TIME) each time it runs `deploy`
VERSION = $(shell cat version.txt)

.PHONY: build
build:
	@echo "==> Building locally"
	go build -o $(NAME)

.PHONY: test
test: build
	@echo "==> Running image locally"
	./$(NAME) --env local

.PHONY: docker
docker:
	@echo "==> Building for linux/amd64"
	env GOOS=linux GOARCH=amd64 go build -o $(NAME)-linux-amd64
	@echo "==> Building image $(NAME)"
	docker build . -t $(NAME)
	@echo "==> Removing exec $(NAME)"
	@rm -fv $(NAME)-linux-amd64

.PHONY: push
push: docker
	@echo "==> Tagging and pushing image as $(TAG)"
	docker tag $(NAME) $(REPO)/$(NAME)
	docker tag $(NAME) $(REPO)/$(NAME):$(TAG)
	docker push $(REPO)/$(NAME)
	docker push $(REPO)/$(NAME):$(TAG)
	@echo "==> Saving $(TAG) into version.txt"
	$(shell echo $(TAG) > version.txt)

.PHONY: deploy
deploy: push
	kubectl set image deployment/$(NAME) echo=$(REPO)/$(NAME):$(VERSION)
