FROM golang:1.21

WORKDIR /app

COPY . .

RUN go build -o goapp cmd/*

EXPOSE 8080

CMD [ "./goapp"]