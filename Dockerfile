# Builder
FROM golang:latest
WORKDIR /usr/src/app
COPY . .
RUN go build

# Service
FROM ubuntu:latest
WORKDIR /usr/bin
COPY --from=0 /usr/src/app/ttrack_api .
EXPOSE 3000/tcp
ENTRYPOINT ["ttrack_api"]
