FROM golang:1.26
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/portfolio .

WORKDIR /app
EXPOSE 8112
CMD ["portfolio"]
