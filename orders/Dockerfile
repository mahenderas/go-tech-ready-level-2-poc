# syntax=docker/dockerfile:1

FROM golang:1.24.4-alpine AS builder
WORKDIR /app/orders
COPY ./orders/go.mod ./orders/go.sum ./
RUN go mod download
COPY ./orders .
RUN go build -o orders .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/orders/orders .
COPY /wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh
EXPOSE 8002
EXPOSE 8080
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432
ARG DEPLOY_ENV=local
ENV DEPLOY_ENV=${DEPLOY_ENV}

# Use wait-for-postgres.sh for local, run binary directly for GCP
CMD if [ "$DEPLOY_ENV" = "gcp" ]; then ./orders; else /wait-for-postgres.sh $POSTGRES_HOST $POSTGRES_PORT ./orders; fi
