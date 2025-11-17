FROM oven/bun:1

WORKDIR /app
COPY . .

RUN apt-get update && \
    apt-get install -y wget git build-essential && \
    wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz && \
    rm go1.25.3.linux-amd64.tar.gz

ENV GOPATH="/root/go"
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"


WORKDIR /app/bin

RUN mkdir -p /root/go/bin

RUN mv marki_x86_64 marki && \
    mv marki /usr/local/bin/


RUN which marki
RUN marki --help || true
RUN bun --version && go version

WORKDIR /app
RUN bun install

EXPOSE 8080

CMD ["bun", "run", "index.ts"]
