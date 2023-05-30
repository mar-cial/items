FROM golang as build

WORKDIR /go/src/items
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/items

FROM gcr.io/distroless/base-debian11

COPY --from=build /go/bin/items /
CMD ["/items"]