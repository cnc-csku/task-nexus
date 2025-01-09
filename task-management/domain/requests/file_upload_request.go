package requests

type GeneratePutPresignedUrlRequest struct {
	Key      string `json:"key"`
	IsPublic bool   `json:"is_public"`
}
