package database_Conn

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	"github.com/go-redis/redis/v8"
)

// RedisDB represents the Redis database
type RedisDB struct {
	Client *redis.Client
	Key    []byte // Encryption key
}

// NewRedisDB creates a new instance of RedisDB
func NewRedisDB(key string) *RedisDB {
	return &RedisDB{
		Client: redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0}),
		Key:    []byte(key),
	}
}

// FlushAll deletes all keys from all Redis databases
func (r *RedisDB) FlushAll() error {
	ctx := context.Background()
	return r.Client.FlushAll(ctx).Err()
}

// StorePatient stores patient data securely in Redis
func (rdb *RedisDB) StorePatient(patientID string, patientData interface{}) error {
	// Convert patient data to JSON
	dataJSON, err := json.Marshal(patientData)
	if err != nil {
		return err
	}

	// Calculate data hash for integrity checking
	dataHash := calculateHash(dataJSON)

	// Encrypt patient data
	encryptedData, err := rdb.encrypt(dataJSON)
	if err != nil {
		return err
	}

	// Store encrypted data and hash in Redis
	pipeline := rdb.Client.TxPipeline()
	pipeline.Set(context.Background(), patientID+":data", encryptedData, 0)
	pipeline.Set(context.Background(), patientID+":hash", dataHash, 0)
	_, err = pipeline.Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// GetPatient retrieves and decrypts patient data from Redis
func (rdb *RedisDB) GetPatient(patientID string, patientData interface{}) error {
	// Retrieve encrypted data and hash from Redis
	pipeline := rdb.Client.TxPipeline()
	dataCmd := pipeline.Get(context.Background(), patientID+":data")
	hashCmd := pipeline.Get(context.Background(), patientID+":hash")
	_, err := pipeline.Exec(context.Background())
	if err != nil {
		return err
	}

	// Check if data and hash exist
	encryptedData, err := dataCmd.Result()
	if err != nil {
		return err
	}
	dataHash, err := hashCmd.Result()
	if err != nil {
		return err
	}

	// Decrypt patient data
	decryptedData, err := rdb.decrypt(encryptedData)
	if err != nil {
		return err
	}

	// Verify data integrity
	if !verifyHash([]byte(decryptedData), dataHash) {
		return errors.New("data integrity check failed")
	}

	// Unmarshal JSON
	err = json.Unmarshal([]byte(decryptedData), &patientData)
	if err != nil {
		return err
	}

	return nil
}

// calculateHash calculates hash of data
func calculateHash(data []byte) string {
	// Implement your hash calculation algorithm here
	// This is just a placeholder
	return "hash" // Replace with actual hash value
}

// verifyHash verifies integrity of data
func verifyHash(data []byte, hash string) bool {
	// Implement your hash verification algorithm here
	// This is just a placeholder
	return true // Replace with actual hash verification logic
}

// encrypt encrypts data using AES block cipher
func (rdb *RedisDB) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(rdb.Key)
	if err != nil {
		return "", err
	}

	// Pad data to be a multiple of block size
	data = padData(data, aes.BlockSize)

	// Create a new cipher block mode for AES encryption
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], data)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// decrypt decrypts data using AES block cipher
func (rdb *RedisDB) decrypt(encryptedData string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(rdb.Key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Create a new cipher block mode for AES decryption
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	// Remove padding
	cipherText = unpadData(cipherText)

	return string(cipherText), nil
}

// padData adds PKCS#7 padding to data
func padData(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// unpadData removes PKCS#7 padding from data
func unpadData(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
