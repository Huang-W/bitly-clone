FROM golang:1.14-alpine
RUN mkdir /app
RUN mkdir /app/src
ADD ./src /app/src/
WORKDIR /app
ENV GOPATH /app
RUN go install datastore-worker
RUN rm -rf /app/src
CMD ["/app/bin/datastore-worker"]
