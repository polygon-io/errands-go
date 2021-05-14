FROM golang:1.16

WORKDIR /src/
COPY . /src/
RUN cd cmd/echo && go build -o /bin/echo-errand

ENTRYPOINT ["/bin/echo-errand"]