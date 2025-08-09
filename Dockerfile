# Multi-stage build for sb29guard
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/sb29guard ./cmd/sb29guard

# Minimal runtime with CA certs
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/sb29guard /app/sb29guard
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/sb29guard"]
CMD ["--help"]
