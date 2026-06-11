FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/eldermere ./cmd/eldermere

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out/eldermere /app/eldermere
EXPOSE 8080
ENTRYPOINT ["/app/eldermere"]
