# building go executable in own stage
FROM golang:1.18

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN go build -o bot .

ENTRYPOINT ["/app/bot"]