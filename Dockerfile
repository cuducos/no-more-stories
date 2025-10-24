FROM golang:1.25-trixie AS build
ENV GOEXPERIMENT=jsonv2
WORKDIR /nms
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /usr/bin/nms && go clean cache

FROM debian:trixie-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*
COPY --from=build /usr/bin/nms /usr/bin/nms
CMD ["nms"]
