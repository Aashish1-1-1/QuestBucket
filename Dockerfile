FROM golang:1.22

WORKDIR /usr/src/app

COPY . .

CMD ["go","run","main.go"]
