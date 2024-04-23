FROM golang:1.22.1 AS builder

WORKDIR /app
COPY . .

# Install MinGW-w64 compiler
RUN apt-get update && apt-get install -y gcc-mingw-w64

# Add MinGW-w64 compiler to PATH
ENV PATH="/usr/x86_64-w64-mingw32/bin:${PATH}"

# Build the application
RUN go mod download && \
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o ./app_windows.exe ./cmd

FROM ubuntu:latest

WORKDIR /app
COPY --from=builder /app/app_windows.exe /app/app_windows.exe
COPY config /app/config
COPY web /app/web
VOLUME /app/db

CMD ["/app/app_windows.exe"]
