FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/xinzhou-science-platforms ./cmd/server
FROM debian:bookworm-slim
COPY --from=build /out/xinzhou-science-platforms /usr/local/bin/xinzhou-science-platforms
COPY migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/xinzhou-science-platforms"]
