FROM golang

WORKDIR /src/github.com/san-lab
RUN cd /src/github.com/san-lab
RUN git clone https://github.com/san-lab/secretsplitcli.git && \
 git clone https://github.com/san-lab/udpsigner.git 

WORKDIR  /src/github.com/san-lab/udpsigner
COPY ./*.json ./
ENV GOPATH=/
WORKDIR  /src/github.com/san-lab/udpsigner
RUN go get golang.org/x/crypto/sha3
RUN go get github.com/schollz/peerdiscovery
RUN go get github.com/google/uuid
RUN go get go.dedis.ch/kyber/pairing
RUN go get github.com/manifoldco/promptui
RUN go get golang.org/x/term

RUN go build

CMD ["/src/github.com/san-lab/udpsigner/udpsigner"]
