package processor

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

func Init() {
	pipeline.Register(&LLMExtract{})
	pipeline.Register(&LLMReview{})
	pipeline.Register(&OCR{})
	pipeline.Register(&Face{})
	pipeline.Register(&Classifier{})
	pipeline.Register(&Embedding{})
	pipeline.Register(&ContextLoader{})
	pipeline.Register(&EntityAlign{})
	pipeline.Register(&ImageAssess{})
	pipeline.Register(&SubPipeline{})
	pipeline.Register(&Router{})
	pipeline.Register(&Summarizer{})
}

func parseJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

var urlRe = regexp.MustCompile(`https?://\S+`)

func buildUserMessage(pctx *pipeline.ProcessorContext, text string) llm.Message {
	if pctx.Document.Type == "image" {
		imageURL := pctx.Document.SourceURL
		if imageURL == "" {
			imageURL = urlRe.FindString(pctx.Document.Content)
		}
		if imageURL != "" && pctx.SourceImageURL == "" {
			pctx.SourceImageURL = imageURL
		}

		imageRef := imageURL
		if imageRef == "" {
			imageRef = pctx.ImageBase64
		}
		if imageRef == "" {
			return llm.Message{Role: "user", Content: text}
		}

		userText := strings.TrimSpace(urlRe.ReplaceAllString(pctx.Document.Content, ""))
		if text != "" && text != userText {
			if userText != "" {
				userText = text + "\n\n用户备注：" + userText
			} else {
				userText = text
			}
		}
		parts := []llm.ContentPart{
			{Type: "image_url", ImageURL: &llm.ImageURL{URL: imageRef}},
		}
		if userText != "" {
			parts = append(parts, llm.ContentPart{Type: "text", Text: userText})
		}
		return llm.Message{Role: "user", Content: parts}
	}
	return llm.Message{Role: "user", Content: text}
}

func doChat(ctx context.Context, pctx *pipeline.ProcessorContext, req llm.Request) (*llm.Response, error) {
	if o := pctx.LLMOverride; o != nil {
		if req.Provider == "" {
			req.Provider = o.ProviderType
		}
		if req.Model == "" {
			req.Model = o.Model
		}
		return pctx.LLM.ChatWithCredential(ctx, req, o.ProviderType, o.APIKey, o.BaseURL)
	}
	return pctx.LLM.Chat(ctx, req)
}

func doEmbedding(ctx context.Context, pctx *pipeline.ProcessorContext, req llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	if o := pctx.LLMOverride; o != nil {
		if req.Provider == "" {
			req.Provider = o.ProviderType
		}
		if req.Model == "" {
			req.Model = o.Model
		}
		return pctx.LLM.EmbeddingWithCredential(ctx, req, o.ProviderType, o.APIKey, o.BaseURL)
	}
	return pctx.LLM.Embedding(ctx, req)
}
