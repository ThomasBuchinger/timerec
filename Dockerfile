FROM golang:1.17 as builder

ENV CGO_ENABLED=0
WORKDIR src/github.com/thomasbuchinger/timerec
COPY go.mod go.sum ./
RUN go mod download

RUN echo '{ "user": {}, "templates": [], "jobs": {}, "records": [] }' > /db.yaml

COPY api api
COPY cmd cmd
COPY internal internal
COPY Makefile README.md ./

RUN make test
RUN go install ./...

FROM golang:1.17 as debug
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=builder /db.yaml /
COPY --from=builder /go/bin/* /
CMD ["/timerec-server"]

FROM scratch AS app
EXPOSE 8080
USER 1000
COPY --from=builder /db.yaml /
COPY --from=builder /go/bin/* /
CMD ["/timerec-server"]