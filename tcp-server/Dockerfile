# syntax=docker/dockerfile:1

FROM golang:1.22.1 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /tcp-server



FROM gcr.io/distroless/static-debian11 AS release-stage

USER nonroot:nonroot

WORKDIR /app

COPY --from=build-stage /tcp-server /app/tcp-server

EXPOSE 7890

ENTRYPOINT ["/app/tcp-server"]
