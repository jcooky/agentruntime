FROM golang:1.24-alpine3.21 AS builder

RUN apk add --no-cache make git binutils

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make bin/agentruntime

FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates nodejs npm

WORKDIR /app
ENV HOME=/app
RUN mkdir /app/agents && chown -R 1000:1000 /app
USER 1000:1000
COPY --from=builder --chown=1000:1000 /app/README.md /app/README.md
COPY --from=builder --chown=1000:1000 /app/examples /app/examples
COPY --from=builder --chown=1000:1000 /app/bin/agentruntime* /app/

ENTRYPOINT ["./agentruntime"]