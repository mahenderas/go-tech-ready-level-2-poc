# syntax=docker/dockerfile:1

FROM golang:1.24.4-alpine AS build
WORKDIR /app/authentication
COPY ./authentication/go.mod ./authentication/go.sum ./
RUN go mod download
COPY ./authentication .
RUN go mod tidy
RUN go build -o authentication .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/authentication/authentication .
COPY /wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh
EXPOSE 8004
EXPOSE 8080
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432
ARG DEPLOY_ENV=local
ENV DEPLOY_ENV=${DEPLOY_ENV}
CMD if [ "$DEPLOY_ENV" = "gcp" ]; then ./authentication; else /wait-for-postgres.sh $POSTGRES_HOST $POSTGRES_PORT ./authentication; fi
