FROM golang:1.26-alpine AS build

WORKDIR /src
COPY apps/server/go.mod apps/server/go.sum ./apps/server/
WORKDIR /src/apps/server
RUN go mod download
WORKDIR /src
COPY apps/server ./apps/server
COPY content-packs ./content-packs
WORKDIR /src/apps/server
RUN go build -o /out/eldermere ./cmd/eldermere

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out/eldermere /app/eldermere
COPY --from=build /src/content-packs /app/content-packs
EXPOSE 8080
ENTRYPOINT ["/app/eldermere"]
