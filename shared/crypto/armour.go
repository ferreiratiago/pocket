package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	// Encryption params
	kdf          = "scrypt"
	randBz       = 16
	AESNonceSize = 12
	// Scrypt params
	n    = 32768 // CPU/memory cost param; power of 2 greater than 1
	r    = 8     // r * p < 2³⁰
	p    = 1     // r * p < 2³⁰
	klen = 32    // bytes
	// Sec param
	secParam = 12
)

// Errors
var (
	ErrorWrongPassphrase = errors.New("Can't decrypt private key: wrong passphrase")
)

// Armoured Private Key struct with fields to unarmour it later
type armouredKey struct {
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	SecParam   string `json:"secparam"`
	Hint       string `json:"hint"`
	CipherText string `json:"ciphertext"`
}

// Generate new armoured private key struct with parameters for unarmouring
func newArmouredKey(kdf, salt, hint, cipherText string) armouredKey {
	return armouredKey{
		Kdf:        kdf,
		Salt:       salt,
		SecParam:   strconv.Itoa(secParam),
		Hint:       hint,
		CipherText: cipherText,
	}
}

// Encrypt the given privKey with the passphrase, armour it by encoding the encrypted
// []byte into base64, and convert into a json string with the parameters for unarmouring
func encryptArmourPrivKey(privKey PrivateKey, passphrase, hint string) (string, error) {
	// Encrypt privKey usign AES-256 GCM Cipher
	saltBz, encBz, err := encryptPrivKey(privKey, passphrase)
	if err != nil {
		return "", err
	}

	// Armour encrypted bytes by encoding into Base64
	armourStr := base64.RawStdEncoding.EncodeToString(encBz)

	// Create ArmouredKey object so can unarmour later
	armoured := newArmouredKey(kdf, hex.EncodeToString(saltBz), hint, armourStr)

	// Encode armoured struct into []byte
	js, err := json.Marshal(armoured)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// Encrypt the given privKey with the passphrase using a randomly
// generated salt and the AES-256 GCM cipher
func encryptPrivKey(privKey PrivateKey, passphrase string) (saltBz, encBz []byte, err error) {
	// Get random bytes for salt
	saltBz = randBytes(randBz)

	// Derive key for encryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	encryptionKey, err := scrypt.Key([]byte(passphrase), saltBz, n, r, p, klen)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt using AES
	privKeyHexString := privKey.String()
	encBz, err = encryptAESGCM(encryptionKey, []byte(privKeyHexString))
	if err != nil {
		return nil, nil, err
	}

	return saltBz, encBz, nil
}

// Unarmor and decrypt the private key using the passphrase provided
func unarmourDecryptPrivKey(armourStr, passphrase string) (privKey PrivateKey, err error) {
	// Decode armourStr back into ArmouredKey struct
	ak := armouredKey{}
	err = json.Unmarshal([]byte(armourStr), &ak)
	if err != nil {
		return nil, err
	}

	// Check the ArmouredKey for the correct parameters on kdf and salt
	if ak.Kdf != kdf {
		return nil, fmt.Errorf("Unrecognized KDF type: %v", ak.Kdf)
	}
	if ak.Salt == "" {
		return nil, fmt.Errorf("Missing salt bytes")
	}

	// Decoding the salt
	saltBz, err := hex.DecodeString(ak.Salt)
	if err != nil {
		return nil, fmt.Errorf("Error decoding salt: %v", err.Error())
	}

	// Decoding the "armoured" ciphertext stored in base64
	encBz, err := base64.RawStdEncoding.DecodeString(ak.CipherText)
	if err != nil {
		return nil, fmt.Errorf("Error decoding ciphertext: %v", err.Error())
	}

	// Decrypt the actual privkey with the parameters
	privKey, err = decryptPrivKey(saltBz, encBz, passphrase)

	return privKey, err
}

// Decrypt the AES-256 GCM encrypted bytes using the passphrase given
func decryptPrivKey(saltBz, encBz []byte, passphrase string) (PrivateKey, error) {
	// Derive key for decryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	encryptionKey, err := scrypt.Key([]byte(passphrase), saltBz, n, r, p, klen)
	if err != nil {
		return nil, err
	}

	// Decrypt using AES
	privKeyRawHexBz, err := decryptAESGCM(encryptionKey, encBz)
	if err != nil {
		return nil, err
	}
	bz, err := hex.DecodeString(string(privKeyRawHexBz))
	if err != nil {
		return nil, err
	}

	// Get private key from decrypted bytes
	pk, err := NewPrivateKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// Encrypt using AES-256 GCM Cipher
func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := key[:AESNonceSize]
	encBz := gcm.Seal(nil, nonce, plaintext, nil)
	return encBz, nil
}

// Decrypt using AES-256 GCM Cipher
func decryptAESGCM(key, encBz []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := key[:AESNonceSize]
	result, err := gcm.Open(nil, nonce, encBz, nil)
	if err != nil && strings.Contains(err.Error(), "authentication failed") {
		return nil, ErrorWrongPassphrase
	} else if err != nil {
		return nil, fmt.Errorf("Can't Decrypt Using AES : %s \n", err)
	}
	return result, nil
}

// Use OS randomness
func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
