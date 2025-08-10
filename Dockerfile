# Multi-stage build for sb29guard
# ---> that means vulnerabilities for this top container aren't a big deal
# ---> ...it only runs for a few seconds
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=dirty
ARG DATE
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X 'github.com/RiceC-at-MasonHS/SB29-guard/cmd/sb29guard.version=${VERSION}' -X 'github.com/RiceC-at-MasonHS/SB29-guard/cmd/sb29guard.commit=${COMMIT}' -X 'github.com/RiceC-at-MasonHS/SB29-guard/cmd/sb29guard.date=${DATE}'" -o /out/sb29guard ./cmd/sb29guard

# Minimal runtime with CA certs
# Runtime: distroless base. Optionally pin by digest for reproducibility and to appease scanners:
# FROM gcr.io/distroless/base-debian12@sha256:<digest>
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/sb29guard /app/sb29guard
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/sb29guard"]
CMD ["--help"]
