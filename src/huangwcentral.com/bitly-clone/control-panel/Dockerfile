FROM golang:1.14-alpine
EXPOSE 3000
ENV PORT 3000
RUN mkdir /app
RUN mkdir /app/src
ADD ./src /app/src/
WORKDIR /app
ENV GOPATH /app
RUN go install control-panel
RUN rm -rf /app/src
CMD ["/app/bin/control-panel"]
