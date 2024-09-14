FROM golang:1.22-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN mkdir static
COPY static/style.css ./static

RUN go build -o ./voting_app

# Deploy the application binary into a lean image
FROM alpine:3 AS release-stage
WORKDIR /app
# This is what this will be using, but put it behind a reverse proxy 
# for a whole boat load of reasons
EXPOSE 39293

COPY --from=build-stage /app ./
# This will be filled by a volume later
RUN mkdir ./data

ENTRYPOINT ["/app/voting_app"]

