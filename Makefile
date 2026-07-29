run:
	go run main.go serve 8080 ./config/.env

docker:
	docker build -t portfolio . && docker run --rm \
		-p 8112:8112 \
                -v $(CURDIR)/config:/app/config \
                -v $(CURDIR)/data:/app/data \
		portfolio
                

