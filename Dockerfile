FROM golang:1.22-alpine as builder

WORKDIR /app
RUN apk add --no-cache make nodejs npm git

COPY . .
RUN make install
COPY . /app
RUN make build
RUN ls -l /
RUN ls -l /app 

FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /
COPY --from=builder /app/dreampicai /dreampicai

EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT [ "./dreampicai" ]