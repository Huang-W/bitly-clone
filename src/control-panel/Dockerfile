FROM golang:latest
EXPOSE 3000
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV GOPATH /app
RUN cd /app/src/control-panel ; go install
RUN rm -f /app/Dockerfile
RUN rm -f /app/Makefile
RUN rm -f /app/docker.sh
RUN rm -rf /app/src
CMD ["/app/bin/control-panel"]
