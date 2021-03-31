FROM golang

WORKDIR  /src/github.com/san-lab/udpsigner
COPY ./*.json ./
RUN go get github.com/san-lab/udpsigner
CMD ["/go/bin/udpsigner"]
