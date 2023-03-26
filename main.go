package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var dbConn *gorm.DB

func main() {
	// Parse command-line arguments
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: ./app [call|serve]")
		return
	}

	// Initialize configuration
	initConfig()

	// Initialize database
	err := initDatabase()
	if err != nil {
		fmt.Println("Failed to connect to database", err)
		return
	}

	// Check command-line arguments
	command := args[1]
	switch command {
	case "call":
		// Check for keyword argument
		if len(args) < 3 {
			fmt.Println("Usage: ./app call <keyword>")
			return
		}
		keyword := args[2]

		// Call OpenAI API and log the request
		requestId := uuid.New()

		// Print the UUID
		response, err := callOpenAIAndLog(keyword, requestId.String())
		if err != nil {
			fmt.Printf("Failed to call OpenAI API: %v\n", err)
			return
		}

		// Print response
		fmt.Println(string(response))

	case "serve":
		// Initialize router
		router := initRouter()

		// Start HTTP server
		port := viper.GetString("server.port")
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			return
		}

	default:
		fmt.Println("Usage: ./app [call|serve]")
	}
}

func initConfig() {
	fmt.Println(os.Getenv("CONFIG_FILE"))
	// If a CONFIG_FILE environment variable is set, read configuration from that file
	if configPath := os.Getenv("CONFIG_FILE"); configPath != "" {
		viper.SetConfigFile(configPath)

		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Failed to read configuration file %s: %v\n", configPath, err)
			os.Exit(1)
		}
	}

	// Set default values for configuration properties
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "3306")
	viper.SetDefault("db.username", "root")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.database", "mydb")

	// Set configuration file search paths and file name
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Attempt to read configuration from file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found; using default configuration values.")
		} else {
			fmt.Printf("Failed to read configuration file: %v\n", err)
			os.Exit(1)
		}
	}

	// Bind environment variables to configuration properties
	viper.AutomaticEnv()

	// Print the final configuration
	fmt.Println("Using the following configuration:")
	fmt.Printf("Server port: %s\n", viper.GetString("server.port"))
	fmt.Printf("Database host: %s\n", viper.GetString("db.host"))
	fmt.Printf("Database port: %s\n", viper.GetString("db.port"))
	fmt.Printf("Database username: %s\n", viper.GetString("db.username"))
	fmt.Printf("Database password: %s\n", viper.GetString("db.password"))
	fmt.Printf("Database name: %s\n", viper.GetString("db.database"))
}

func initDatabase() error {
	// Read database configuration from the config file
	dbConfig := viper.GetStringMapString("database")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig["host"], dbConfig["port"], dbConfig["user"], dbConfig["password"], dbConfig["dbname"])

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Perform database migrations
	err = Up(db)
	if err != nil {
		return err
	}

	// Save the database connection to the global variable
	dbConn = db

	return nil
}

func initRouter() *mux.Router {
	// Initialize router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/api/openai", openAIHandler)

	return router
}

func getIpAddress(req *http.Request) string {
	ipAddress := req.Header.Get("X-Real-IP")
	if ipAddress == "" {
		ipAddress = req.Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = req.RemoteAddr
		}
	}
	return ipAddress
}

type RequestBody struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func callOpenAIAndLog(keyword string, requestId string) ([]byte, error) {
	start := time.Now()

	if requestId == "" {
		requestId = uuid.New().String()
	}

	// Load OpenAI API configuration from Viper configuration
	openaiUrl := viper.GetString("openai.url")
	openaiModel := viper.GetString("openai.model")
	openaiApiKey := viper.GetString("openai.api_key")
	orgApiKey := viper.GetString("openai.org_key")
	openaiSystem := viper.GetString("openai.system")
	openaiTemp := viper.GetFloat64("openai.temp")
	openaiMaxToken := viper.GetInt("openai.max_token")

	// Construct the request body
	requestBody := RequestBody{
		Model: openaiModel,
		Messages: []Message{
			{
				Role:    "system",
				Content: openaiSystem,
			},
			{
				Role:    "user",
				Content: keyword,
			},
		},
		MaxTokens:   openaiMaxToken,
		Temperature: float32(openaiTemp),
	}

	requestBodyJSON, err := json.Marshal(requestBody)

	// Send the request to OpenAI API
	req, err := http.NewRequest("POST", openaiUrl, strings.NewReader(string(requestBodyJSON)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiApiKey))
	req.Header.Set("OpenAI-Organization", fmt.Sprintf("%s", orgApiKey))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log the request and response
	requestHeaders := fmt.Sprintf("%v", req.Header)
	responseHeaders := fmt.Sprintf("%v", resp.Header)
	userAgent := req.UserAgent()
	ipAddress := getIpAddress(req)

	err = logRequest(
		req.Method,
		req.URL.String(),
		string(requestBodyJSON),
		requestHeaders,
		responseHeaders,
		string(responseBody),
		resp.StatusCode,
		"",
		userAgent,
		ipAddress,
		time.Since(start),
		start,
		time.Now(),
		req.ContentLength,
		resp.ContentLength,
		requestId,
	)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func logRequest(httpMethod string, requestUrl string, requestBody string, requestHeaders string, responseHeaders string, responseBody string, statusCode int, errorMessage string, userAgent string, ipAddress string, duration time.Duration, requestTimestamp time.Time, responseTimestamp time.Time, requestSize int64, responseSize int64, requestId string) error {
	// Create a new Log struct with the provided data
	log := Log{
		HttpMethod:        httpMethod,
		RequestUrl:        requestUrl,
		RequestBody:       requestBody,
		RequestHeaders:    requestHeaders,
		ResponseHeaders:   responseHeaders,
		ResponseBody:      responseBody,
		StatusCode:        statusCode,
		ErrorMessage:      errorMessage,
		UserAgent:         userAgent,
		IPAddress:         ipAddress,
		Duration:          duration,
		RequestTimestamp:  requestTimestamp,
		ResponseTimestamp: responseTimestamp,
		RequestSize:       requestSize,
		ResponseSize:      responseSize,
		RequestId:         requestId,
	}

	// Save the Log struct to the database
	result := dbConn.Create(&log)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func openAIHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the keyword from the request query string
	keyword := r.URL.Query().Get("keyword")
	requestId := r.URL.Query().Get("requestId")
	fmt.Printf("Call openai, keyword:%s\n", keyword)
	// Call the callOpenAIAndLog function with the keyword
	response, err := callOpenAIAndLog(keyword, requestId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header and write the response to the client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		return
	}
}
