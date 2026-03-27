package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIRequest is the incoming request from the frontend.
type AIRequest struct {
	Message string `json:"message"`
	Context string `json:"context"` // Recent terminal output for context
}

// AIResponse is the response sent back to the frontend.
type AIResponse struct {
	Reply string `json:"reply"`
	Error string `json:"error,omitempty"`
}

// GeminiRequest models the Gemini API request payload.
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	SystemInstruction *GeminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig  map[string]interface{} `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse models the Gemini API response.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

const aiSystemPrompt = `You are ZeroExec AI — a senior systems engineer assistant embedded in a secure browser-based terminal gateway. Your role:

1. Help users with shell commands (PowerShell, cmd, bash).
2. Explain terminal output, errors, and system behavior.
3. Suggest safe, idiomatic commands for the user's task.
4. Warn about destructive or risky operations.
5. Be concise — respond in 1-3 short paragraphs max. Use code blocks for commands.
6. If terminal context is provided, analyze it to give contextual answers.

You are NOT a general chatbot. Stay focused on terminal, DevOps, and systems topics.`

// AIHandler handles the /ai/chat endpoint.
type AIHandler struct {
	apiKey     string
	model      string
	middleware *Middleware
	httpClient *http.Client
}

func NewAIHandler(cfg *Config, mw *Middleware) *AIHandler {
	model := cfg.AI.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &AIHandler{
		apiKey:     cfg.AI.APIKey,
		model:      model,
		middleware: mw,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *AIHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Auth
	_, err := h.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check API key
	if h.apiKey == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIResponse{
			Error: "AI Assistant not configured. Set ai.api_key in config.yaml or VT_AI_KEY environment variable.",
		})
		return
	}

	// Parse request
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, "Empty message", http.StatusBadRequest)
		return
	}

	// Build Gemini request
	userContent := req.Message
	if req.Context != "" {
		userContent = fmt.Sprintf("Terminal context (recent output):\n```\n%s\n```\n\nUser question: %s", req.Context, req.Message)
	}

	geminiReq := GeminiRequest{
		SystemInstruction: &GeminiContent{
			Parts: []GeminiPart{{Text: aiSystemPrompt}},
		},
		Contents: []GeminiContent{
			{
				Role:  "user",
				Parts: []GeminiPart{{Text: userContent}},
			},
		},
		GenerationConfig: map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 1024,
		},
	}

	body, _ := json.Marshal(geminiReq)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", h.model, h.apiKey)

	resp, err := h.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIResponse{Error: "Failed to reach AI service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIResponse{Error: "Failed to parse AI response"})
		return
	}

	if geminiResp.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIResponse{Error: geminiResp.Error.Message})
		return
	}

	reply := ""
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		reply = geminiResp.Candidates[0].Content.Parts[0].Text
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AIResponse{Reply: reply})
}
