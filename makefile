NAME=$(shell basename "$(PWD)")
ORG=byuoitav
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

ifeq ($(BRANCH),HEAD)
BRANCH := $(shell echo $(CIRCLE_BRANCH))
endif

#docker
UNAME=$(shell echo $(DOCKER_USERNAME))
PASSW=$(shell echo $(DOCKER_PASSWORD))

# detect OS
UNAME_S := $(shell uname -s)

docker: $(NAME)-bin
	docker build --build-arg NAME=$(NAME) -t $(ORG)/$(NAME):$(BRANCH) . 
	docker login -u $(UNAME) -p $(PASSW)
	docker push $(ORG)/$(NAME):$(BRANCH)

$(NAME)-bin:
	env GOOS=linux GGO_ENABLED=0 go build -o $(NAME)-bin -v

clean:
	go clean
	rm $(NAME)-bin
