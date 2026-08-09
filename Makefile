IMAGE_NAME ?= portfolio
HOST_PORT ?= 8112
CONTAINER_PORT ?= 8112

.PHONY: run docker

run:
	go run .

docker:
	docker build -t $(IMAGE_NAME) .
	docker run --rm \
		-p $(HOST_PORT):$(CONTAINER_PORT) \
		-v $(CURDIR)/config:/app/config \
		-v $(CURDIR)/data:/app/data \
		$(IMAGE_NAME)
