# syntax=docker/dockerfile:1
FROM golang:alpine as builder
RUN apk --no-cache add tzdata
WORKDIR /home/dominic/GolandProjects/Server
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix go -o app .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /bin
COPY --from=builder /home/dominic/GolandProjects/Server/app/ .
ENV TZ=America/New_York
EXPOSE 8080
CMD ["./app"]