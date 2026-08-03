run:
	go run main.go

docker:
	docker build -t portfolio . && docker run --rm \
		-p 8112:8112 \
                -v $(CURDIR)/config:/app/config \
                -v $(CURDIR)/data:/app/data \
		portfolio
                
