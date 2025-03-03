FROM golang:1.24

WORKDIR /app
COPY .env .
COPY go.mod go.sum ./ 
RUN go mod download

COPY . .
RUN go build -o main .

EXPOSE 8080

CMD ["sh", "-c", "go run migrate/migrate.go && ./main"]
# CMD ["./main"]
