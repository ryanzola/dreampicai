FROM golang:1.22-alpine as builder

WORKDIR /app
RUN apk add --no-cache make nodejs npm
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./
RUN make install
RUN make build
RUN > /app/.env

FROM scratch
COPY --from=builder /app/bin/dreampicai /dreampicai
COPY --from=builder /app/.env .env

EXPOSE 3000
ENTRYPOINT [ "./dreampicai" ]
