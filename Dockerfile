FROM golang:1.22-alpine as builder

WORKDIR /app
RUN apk add --no-cache make nodejs npm

COPY . ./
RUN make install
RUN make build
RUN > /app/.env

FROM scratch
COPY --from=builder /dreampic /dreampic
COPY --from=builder /app/.env .env

EXPOSE 3000
ENTRYPOINT [ "./dreampic" ]