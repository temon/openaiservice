FROM golang:latest

# Set the working directory
WORKDIR /app

# Copy the Go modules file
COPY go.mod .

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

RUN go mod tidy

# Build the binary
RUN go build -o main .

# Set the configuration file path as an environment variable
ENV CONFIG_FILE=/app/config.yaml

# Expose the port
EXPOSE 1414

CMD ["/bin/bash", "-c", "/bin/sleep 10 && ./main serve"]