package entity

type CloudProviderAWS struct {
	AccessKeyID string         `json:"accessKeyId"`
	SecretKey   EncryptedField `json:"secretKey"`
	Region      string         `json:"region,omitempty"`
}
