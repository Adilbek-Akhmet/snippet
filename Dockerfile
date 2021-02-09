FROM golang
RUN mkdir /app
COPY . /app
ENV CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /app/cmd/web
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /app /app
WORKDIR ./app
EXPOSE 4000
CMD ["/app/cmd/web/main"]
