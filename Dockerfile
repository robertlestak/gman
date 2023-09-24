FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/gman cmd/gman/*.go

FROM debian as app

RUN apt-get update -y && apt-get install -y ca-certificates \
    xdg-utils \
    git \
    less \
    groff \
    pandoc \
    curl \
    curl -fsSL https://deb.nodesource.com/setup_current.x | bash - && \
    apt-get install -y nodejs npm \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/gman /bin/gman

ENTRYPOINT ["/bin/gman"]