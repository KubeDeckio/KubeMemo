FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -X github.com/KubeDeckio/KubeMemo/internal/cli.version=${VERSION}" \
    -o /out/kubememo ./cmd/kubememo

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /out/kubememo /kubememo

USER nonroot:nonroot

ENTRYPOINT ["/kubememo"]
CMD ["version", "--output", "json"]
