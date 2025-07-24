# syntax=docker/dockerfile:1

FROM golang:1.24.4-alpine AS builder
WORKDIR /app/payment
COPY ./payment/go.mod ./payment/go.sum ./
RUN go mod download
COPY ./payment .
RUN go build -o payment .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment/payment .
COPY /wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh
EXPOSE 8003
EXPOSE 8080
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432
ARG DEPLOY_ENV=local
ENV DEPLOY_ENV=${DEPLOY_ENV}
CMD if [ "$DEPLOY_ENV" = "gcp" ]; then ./payment; else /wait-for-postgres.sh $POSTGRES_HOST $POSTGRES_PORT ./payment; fi
