FROM golang:1.22 AS build

WORKDIR /cmd

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app -v ./cmd


FROM scratch AS final

WORKDIR /

COPY --from=build /bin/app /app

EXPOSE 8080
EXPOSE 8090

ENTRYPOINT ["/app"]