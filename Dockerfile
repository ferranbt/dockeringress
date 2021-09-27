FROM golang:1.15.5-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor main.go

FROM gcr.io/distroless/base-debian10

COPY --from=build /src/main /usr/bin/dockeringress
ENTRYPOINT ["/usr/bin/dockeringress"]
