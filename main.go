package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Default Configuration values (used if environment variables are not set)
const (
	defaultVLLMAPIURL = "http://localhost:8000"
	defaultVLLMModel  = "ibm-granite/granite-3.2-8b-instruct"
)

// Structures for the API request and response (OpenAI-compatible)
type APIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"` // Added Temperature field
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

func main() {
	// Get configuration from environment variables or use defaults
	vllmAPIURL := os.Getenv("VLLM_API_URL")
	if vllmAPIURL == "" {
		vllmAPIURL = defaultVLLMAPIURL
	}

	vllmModel := os.Getenv("VLLM_MODEL")
	if vllmModel == "" {
		vllmModel = defaultVLLMModel
	}

	vllmAPIToken := os.Getenv("VLLM_API_TOKEN") // No default for token, it's optional

	// --- Command-line flag parsing ---
	systemPrompt := flag.String("system", "You are a helpful assistant.", "System prompt for the LLM")
	userPromptFlag := flag.String("user", "", "User prompt for the LLM (overrides stdin)")
	temperatureFlag := flag.Float64("temperature", 0.7, "Temperature for LLM generation (e.g., 0.2 for more deterministic, 1.0 for more random)") // Added temperature flag
	flag.Parse()

	var userPrompt string

	if *userPromptFlag != "" {
		userPrompt = *userPromptFlag
	} else {
		// Check if input is being piped
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 { // Check if data is piped
			stdinBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			userPrompt = strings.TrimSpace(string(stdinBytes))
		}
	}

	if userPrompt == "" && len(flag.Args()) > 0 {
		userPrompt = strings.Join(flag.Args(), " ")
	}

	if userPrompt == "" {
		fmt.Fprintln(os.Stderr, "Error: User prompt is required. Provide it via --user flag, pipe, or as a trailing argument.")
		flag.Usage()
		os.Exit(1)
	}

	// --- API Request Logic ---
	messages := []Message{
		{Role: "system", Content: *systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	apiRequest := APIRequest{
		Model:       vllmModel,
		Messages:    messages,
		Temperature: *temperatureFlag,
	}

	jsonData, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}

	// Construct the full API endpoint
	fullAPIURL := vllmAPIURL
	if !strings.HasSuffix(fullAPIURL, "/") {
		fullAPIURL += "/"
	}
	fullAPIURL += "v1/chat/completions"

	req, err := http.NewRequest("POST", fullAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	if vllmAPIToken != "" {
		req.Header.Set("Authorization", "Bearer "+vllmAPIToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request to vLLM API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error: API request failed with status %s: %s\n", resp.Status, string(bodyBytes))
		os.Exit(1)
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding API response: %v\n", err)
		os.Exit(1)
	}

	if len(apiResponse.Choices) > 0 {
		fmt.Println(apiResponse.Choices[0].Message.Content)
	} else {
		fmt.Fprintln(os.Stderr, "No choices returned in API response.")
	}
}
