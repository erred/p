FROM golang:alpine AS build
ARG CGO_ENABLED=0
ARG GOPROXY=https://proxy.golang.org,direct
ARG GOMODCACHE=/go/pkg/mod
ARG GOCACHE=/root/.cache/go-build
WORKDIR /workspace
COPY . .
RUN go build -trimpath -ldflags='-s -w' -o p

FROM gcr.io/distroless/static
COPY --from=build /workspace/p /bin/p
ENTRYPOINT ["/bin/p"]
