FROM golang

WORKDIR  /src/github.com/san-lab
RUN git clone https://github.com/san-lab/udpsigner
WORKDIR /src/github.com/san-lab/udpsigner
COPY ./*.json ./
RUN go build
ENV httpPort "8100"
ENV withHttp "false"
CMD ./udpsigner -httpPort=$httpPort -withHttp=$withHttp 
