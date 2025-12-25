FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/api .
EXPOSE 8080
CMD [ "./api" ]