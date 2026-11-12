package url

//go:generate easyjson -all url.go
type CreateURLRequest struct {
	URL string `json:"url" binding:"required"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}

type CreateURLBatchRequest struct {
	CorrelationID string `json:"correlation_id" binding:"required"`
	OriginalURL   string `json:"original_url" binding:"required"`
}

type CreateURLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type CreateURLBatchRequests []CreateURLBatchRequest

//easyjson:json
type CreateURLBatchResponses []CreateURLBatchResponse
