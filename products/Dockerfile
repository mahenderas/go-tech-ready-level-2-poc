# syntax=docker/dockerfile:1

FROM golang:1.24.4-alpine AS build
WORKDIR /app/products
COPY ./products/go.mod ./products/go.sum ./
RUN go mod download
COPY ./products .
RUN go build -o products .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/products/products .
COPY /wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh
EXPOSE 8001
EXPOSE 8080
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432
ARG DEPLOY_ENV=local
ENV DEPLOY_ENV=${DEPLOY_ENV}
CMD if [ "$DEPLOY_ENV" = "gcp" ]; then ./products; else /wait-for-postgres.sh $POSTGRES_HOST $POSTGRES_PORT ./products; fi
