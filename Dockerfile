FROM golang:1.8

ENV PORT 3000

RUN go get -u github.com/colm2/impressive

EXPOSE 3000

CMD /bin/bash -c "cd /go/bin && ./impressive"