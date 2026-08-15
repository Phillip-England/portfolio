FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/portfolio .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /usr/local/bin/portfolio /usr/local/bin/portfolio
COPY --from=build /src/static ./static
COPY --from=build /src/posts ./posts
EXPOSE 8112
CMD ["portfolio"]
