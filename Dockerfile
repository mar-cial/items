
## build

FROM golang AS build

WORKDIR /go/src/items

COPY go.sum go.mod ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -o /items

RUN go test -v ./...

## deploy

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /items /items

EXPOSE 8000

CMD ["/items"]
