FROM golang:1.21.5-alpine AS builder

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main .

FROM alpine AS production

COPY --from=builder /app .

CMD ["./main"]

