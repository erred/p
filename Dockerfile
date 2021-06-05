FROM golang:alpine AS build
ENV CGO_ENABLED=0
WORKDIR /workspace
COPY . .
RUN go build -trimpath -ldflags='-s -w' -o p

FROM gcr.io/distroless/static
COPY --from=build /workspace/p /bin/p
ENTRYPOINT ["/bin/p"]
