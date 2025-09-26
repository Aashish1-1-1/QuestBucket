FROM golang:1.24.6

WORKDIR /usr/src/app

COPY . .

CMD ["go","run","main.go"]
