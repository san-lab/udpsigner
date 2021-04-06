FROM golang

WORKDIR  /src/github.com/san-lab/udpsigner
COPY ./*.json ./
RUN go get github.com/san-lab/udpsigner@json_output
ENV httpPort "8100"
ENV withHttp "false"
CMD /go/bin/udpsigner -httpPort=$httpPort -withHttp=$withHttp ]
