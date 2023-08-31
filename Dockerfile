# Build stage
FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sr-api

# Final stage
FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /sr-api /sr-api

EXPOSE 8080

CMD ["/sr-api"]
