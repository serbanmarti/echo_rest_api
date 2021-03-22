FROM golang:1.16.2-alpine3.12

# Set the workdir
WORKDIR /app

# Copy module files to run go download first
COPY go.mod .
COPY go.sum .

# Get the project dependencies
RUN go mod download

# ADD the source code
COPY . .

# Build the application executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /app/backend /app/cmd/echo_rest_api/main.go

ENV INST_PORT=80
EXPOSE 80
# Set the entry point for running container
ENTRYPOINT ["/app/backend"]