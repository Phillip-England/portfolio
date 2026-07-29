FROM golang:1.26
WORKDIR /app
RUN apt-get update \
    && apt-get install -y --no-install-recommends poppler-utils \
    && rm -rf /var/lib/apt/lists/*
COPY . .
RUN go build -o /usr/local/bin/portfolio .
EXPOSE 8080
CMD ["portfolio", "serve", "8112", "./config/.env"]
