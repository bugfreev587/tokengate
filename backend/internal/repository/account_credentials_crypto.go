package repository

import (
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	accountCredentialsEncryptedKey = "_encrypted"
	accountCredentialsCipherKey    = "_ciphertext"
	accountCredentialsSchemaKey    = "_schema"
	accountCredentialsSchemaV1     = "account_credentials.v1"
)

func encryptAccountCredentials(encryptor service.SecretEncryptor, credentials map[string]any) (map[string]any, error) {
	if encryptor == nil {
		return nil, fmt.Errorf("account credential encryptor is not configured")
	}
	plaintext, err := json.Marshal(normalizeJSONMap(credentials))
	if err != nil {
		return nil, fmt.Errorf("marshal account credentials: %w", err)
	}
	ciphertext, err := encryptor.Encrypt(string(plaintext))
	if err != nil {
		return nil, fmt.Errorf("encrypt account credentials: %w", err)
	}
	return map[string]any{
		accountCredentialsEncryptedKey: true,
		accountCredentialsSchemaKey:    accountCredentialsSchemaV1,
		accountCredentialsCipherKey:    ciphertext,
	}, nil
}

func decryptAccountCredentials(encryptor service.SecretEncryptor, wrapped map[string]any) (map[string]any, error) {
	if encryptor == nil {
		return nil, fmt.Errorf("account credential encryptor is not configured")
	}
	ciphertext, ok := wrapped[accountCredentialsCipherKey].(string)
	if !ok || ciphertext == "" {
		return nil, fmt.Errorf("encrypted account credentials missing ciphertext")
	}
	plaintext, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt account credentials: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(plaintext), &out); err != nil {
		return nil, fmt.Errorf("unmarshal account credentials: %w", err)
	}
	return normalizeJSONMap(out), nil
}

func isEncryptedAccountCredentials(credentials map[string]any) bool {
	if credentials == nil {
		return false
	}
	encrypted, _ := credentials[accountCredentialsEncryptedKey].(bool)
	schema, _ := credentials[accountCredentialsSchemaKey].(string)
	return encrypted && schema == accountCredentialsSchemaV1
}

func firstSecretEncryptor(encryptors []service.SecretEncryptor) service.SecretEncryptor {
	if len(encryptors) == 0 {
		return nil
	}
	return encryptors[0]
}
