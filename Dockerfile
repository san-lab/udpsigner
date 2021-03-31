FROM golang

WORKDIR /src/github.com/san-lab
RUN cd /src/github.com/san-lab
RUN git clone https://github.com/san-lab/secretsplitcli.git && \
 git clone https://github.com/san-lab/udpsigner.git 

WORKDIR  /src/github.com/san-lab/udpsigner
COPY ./*.json ./
RUN go get github.com/san-lab/udpsigner

CMD ["/src/github.com/san-lab/udpsigner/udpsigner"]
