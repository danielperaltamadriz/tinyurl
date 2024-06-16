FROM alpine AS certs
RUN apk --no-cache add ca-certificates


FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o /app/server ./cmd/api
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o /app/website ./cmd/website


FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server
COPY --from=builder /app/website /website
USER 1000

ENTRYPOINT ["/server"]