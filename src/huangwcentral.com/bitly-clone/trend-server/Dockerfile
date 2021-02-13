FROM golang:1.14-alpine
EXPOSE 3002
ENV PORT 3002
RUN mkdir /app
RUN mkdir /app/src
ADD ./src /app/src/
WORKDIR /app
ENV GOPATH /app
RUN go install trend-server
RUN rm -rf /app/src
CMD ["/app/bin/trend-server"]
