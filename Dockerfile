FROM oven/bun:1

WORKDIR /app
COPY . .

# installing go
RUN apt-get update && \
    apt-get install -y wget git build-essential && \
    wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz && \
    rm go1.25.3.linux-amd64.tar.gz

# Add Go paths so "go install" binaries become runnable
ENV GOPATH="/root/go"
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

# installing marki for markdown conversions
RUN go install github.com/phillip-england/marki@latest

# checking if marki is installed
RUN which marki

# confirm bun and go are installed
RUN bun --version && go version

# installing packages
RUN bun install

# running app
CMD ["bun", "run", "index.ts"]


EXPOSE 8080