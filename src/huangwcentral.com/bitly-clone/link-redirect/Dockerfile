FROM golang:1.14-alpine
EXPOSE 3001
ENV PORT 3001
RUN mkdir /app
RUN mkdir /app/src
ADD ./src /app/src/
WORKDIR /app
ENV GOPATH /app
RUN go install link-redirect
RUN rm -rf /app/src
CMD ["/app/bin/link-redirect"]
