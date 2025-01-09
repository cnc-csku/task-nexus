package responses

type GeneratePutPresignedUrlResponse struct {
	Key          string `json:"key"`
	PresignedUrl string `json:"presigned_url"`
	ExpiredAt    string `json:"expired_at"`
}
