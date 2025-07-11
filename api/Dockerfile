FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o api .

EXPOSE 3000

CMD ["./api"]
