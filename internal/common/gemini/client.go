package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	client *genai.Client
}

type GenerateRequest struct {
	Model             string
	SystemInstruction string
	Prompt            string
	Temperature       float32
	ResponseSchema    *genai.Schema
}

type GenerateResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model := c.client.GenerativeModel(req.Model)
	model.Temperature = &req.Temperature

	if req.SystemInstruction != "" {
		model.SystemInstruction = genai.NewUserContent(genai.Text(req.SystemInstruction))
	}
	if req.ResponseSchema != nil {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = req.ResponseSchema
	}

	resp, err := model.GenerateContent(ctx, genai.Text(req.Prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini API call failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response part type")
	}

	result := &GenerateResponse{Text: string(text)}
	if resp.UsageMetadata != nil {
		result.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		result.CompletionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return result, nil
}

// GenerateJSON calls Generate and unmarshals the JSON response into dest.
func (c *Client) GenerateJSON(ctx context.Context, req GenerateRequest, dest any) (*GenerateResponse, error) {
	resp, err := c.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(resp.Text), dest); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return resp, nil
}
