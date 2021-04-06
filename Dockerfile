FROM golang

WORKDIR /src/github.com/san-lab/udpsigner
COPY . .
RUN go build
ENV httpPort "8100"
ENV withHttp "false"
CMD ./udpsigner -httpPort=$httpPort -withHttp=$withHttp 
