FROM golang:1.22

WORKDIR /usr/src/app

RUN go mod init user

COPY . .

CMD ["go","run","main.go"]
