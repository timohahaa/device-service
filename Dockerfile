from golang:1.21 AS builder

WORKDIR /src
COPY . .

# зависимости
RUN go mod download

# компилируем 
RUN CGO_ENABLED=0 GOOS=linux go build -o ./binary cmd/main.go

###################################################

# main stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /src/binary ./scanner
COPY --from=builder /src/config/config.yaml ./config/config.yaml

# пока оставлю хардкод-директории, чтобы локально запускать и тестить контейнер
RUN mkdir /in
RUN mkdir /out

CMD ./scanner
