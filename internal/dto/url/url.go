package url

//go:generate easyjson -all url.go
type CreateURLRequest struct {
	URL string `json:"url" binding:"required"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}
