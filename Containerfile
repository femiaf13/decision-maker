FROM golang:1.22-bookworm AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN mkdir static
COPY static/style.css ./static
# This will be filled by a volume later
RUN mkdir /app/data

RUN CGO_ENABLED=1 go build -o ./voting_app

# Deploy the application binary into a lean image
FROM  debian:bookworm-slim AS release-stage
WORKDIR /
# This is what this will be using, but put it behind a reverse proxy 
# for a whole boat load of reasons
EXPOSE 39293

COPY --from=build-stage /app/ /

ENTRYPOINT ["/voting_app"]

