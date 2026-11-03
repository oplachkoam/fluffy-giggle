package port

import "context"

// ChatRequest is one call to a chat model: a system prompt, a user message, an
// optional image and whether a JSON object is expected back.
type ChatRequest struct {
	System string
	User   string
	Image  *ImageData
	JSON   bool
}

type ImageData struct {
	Data []byte
	MIME string
}

type ChatResponse struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
}

type Model interface {
	Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// Models are the three endpoints by role. Local and Vision stay inside the
// perimeter; Cloud is reached only through the egress check.
type Models struct {
	Local  Model
	Vision Model
	Cloud  Model
}
