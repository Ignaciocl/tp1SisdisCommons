FROM golang
WORKDIR /app
COPY go.* .
RUN go mod download
COPY .. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /program
ENTRYPOINT ["/program"]
