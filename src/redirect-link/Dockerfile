FROM golang:latest
EXPOSE 3001
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV GOPATH /app
RUN cd /app/src/redirect-link ; go install
RUN rm -f /app/Dockerfile
RUN rm -f /app/Makefile
RUN rm -f /app/docker.sh
RUN rm -rf /app/src
CMD ["/app/bin/redirect-link"]
