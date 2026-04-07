# ---- Build Stage ----
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY main.go .
COPY internal ./internal
COPY resources ./resources

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o /usr/bin/pipeline .

# ---- Final Stage ----
# It assumes --platform=$TARGETPLATFORM by default
# We build from the Python base image so we can run the Health Definition stage
# in this same image.
FROM public.ecr.aws/docker/library/python:3.12-alpine

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/promise-postgresql

COPY --from=builder /usr/bin/pipeline /usr/bin/pipeline
COPY --from=builder /app/resources /resources

ENTRYPOINT ["/usr/bin/pipeline"]
