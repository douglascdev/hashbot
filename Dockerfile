FROM golang:1.24

WORKDIR /usr/src

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY . .
RUN \
go mod download && \
go build -o /usr/bin/hashbot



EXPOSE 8080
CMD ["/usr/bin/hashbot", "main.go", "--cfg", "config.hb"]

