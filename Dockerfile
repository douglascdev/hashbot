FROM golang:1.24

WORKDIR /usr/src

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change

EXPOSE 8080
CMD ["go", "run", "main.go", "-cfg", "config.hb"]

