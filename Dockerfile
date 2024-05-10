FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /sr-api

FROM gcr.io/distroless/static

COPY --from=build /sr-api /sr-api

USER nonroot:nonroot

EXPOSE 8080

CMD ["/sr-api"]
