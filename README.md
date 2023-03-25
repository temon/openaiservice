# OpenAI API Client

A simple client for calling the OpenAI API with Go and logging the requests.
## Getting Started
### Prerequisites

- Go 1.16 or higher
- Docker
- Docker Compose

### Installing

    Clone the repository: git clone https://github.com/[USERNAME]/openai-api-client.git
    Move into the project directory: cd openai-api-client
    Install the dependencies: go mod download

### Usage
Running locally

To run the client locally, use the go run command:

        go run main.go serve

By default, the server will listen on port 1414. You can change this by modifying the server.port configuration value in the config.yaml file.

You can make requests to the API by sending a GET request to the /api/openai endpoint with a keyword query parameter:

        http://localhost:1414/api/openai?keyword=[YOUR_KEYWORD]

### Docker
Start the application using Docker Compose:

        docker-compose up

### Contributing

Contributions to this project are welcome! To get started:

- Fork this repository
- Create a new branch: git checkout -b feature/my-new-feature
- Make your changes and commit them: git commit -am 'Add some feature'
- Push the changes to your fork: git push origin feature/my-new-feature
- Create a new pull request

Please make sure your code passes the unit tests and integration tests before submitting a pull request.

### Credit
This amazing project would not have been possible without the contributions of these incredible open-source libraries:
- Golang
- mux
- gorm 
- Viper 