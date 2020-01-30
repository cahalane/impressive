FROM golang:1.13 AS builder

ENV GOBIN=/go/bin

COPY . . 

RUN go get . && \
    CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/impressive

FROM scratch

COPY --from=builder /go/bin/impressive .

ENV PORT 3000

EXPOSE 3000

ENTRYPOINT [ "./impressive" ]