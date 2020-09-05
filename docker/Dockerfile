FROM golang:1.13-alpine AS build


RUN apk add --no-cache git
COPY go.mod /go/src/github.com/falarica/steerd-presto-operator/
WORKDIR /go/src/github.com/falarica/steerd-presto-operator/
RUN go mod download

COPY . /go/src/github.com/falarica/steerd-presto-operator/
RUN go build -o /bin/steerd-presto-operator  cmd/manager/main.go

FROM alpine:3.7
COPY --from=build /bin/steerd-presto-operator /bin/steerd-presto-operator
ENTRYPOINT ["/bin/steerd-presto-operator"]
