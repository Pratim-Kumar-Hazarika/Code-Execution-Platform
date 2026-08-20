FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mini-judge .

FROM alpine:3.21
RUN apk add --no-cache docker-cli ca-certificates
COPY --from=build /mini-judge /mini-judge
EXPOSE 3001
CMD ["/mini-judge"]
