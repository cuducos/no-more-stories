FROM golang:1.24-bookworm AS build
WORKDIR /nms
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /usr/bin/nms

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*
COPY --from=build /usr/bin/nms /usr/bin/nms
CMD ["nms"]
