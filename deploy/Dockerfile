# syntax=docker/dockerfile:1

# --- build stage -----------------------------------------------------------
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/cli ./cmd/cli

# --- final stage ---------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/cli /usr/local/bin/cli

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
