# syntax=docker/dockerfile:1
# builder image
FROM golang:1.18.0-alpine3.15 as builder

RUN mkdir /app
WORKDIR /app
RUN apk add git

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /proxycache 

# generate clean, final image for end users
FROM alpine:3.15
COPY ./.env* ./


COPY --from=builder /proxycache .

CMD [ "./proxycache" ]
