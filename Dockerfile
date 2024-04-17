FROM golang:1.22-alpine as builder

WORKDIR /app
RUN apk add --no-cache make nodejs npm git

COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . /app

RUN make build
RUN ls -la /app 

FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /
COPY --from=builder /app/dreampicai /dreampicai

EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT [ "./dreampicai" ]