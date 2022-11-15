package keys

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	flagRecover = "recover"
)

// createCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creating an encrypted private key and save to <name> file as the key pair identifier",
	Long: `Derive a new private key and encrypt to disk.

Allow users to use BIP39 mnemonic and to secure the mnemonic. Take key ID <name> from user and store key under the <name>.
`,
	Args: cobra.ExactArgs(1),
	RunE: runAddCmd,
}

/*
key type
 - name: the unique ID for the key
 - publickey: the public key
 - privatekey: the private key
 - Address: the address related to the private key
 - mnemonic: mnemonic used to generated key, empty if not saved
*/
type key struct {
	Name       string                  `json:"name"`
	PublicKey  cryptoPocket.PublicKey  `json:"publickey"`
	PrivateKey cryptoPocket.PrivateKey `json:"privatekey"`
	Address    cryptoPocket.Address    `json:"address"`
	Mnemonic   string                  `json:"mnemonic"`
}

// Future updates
// - determine a safer keystore location (team discuss)
// - confirmation from user for overriding existing key
// - implement key phrase intput from user secure keys
func runAddCmd(cmd *cobra.Command, args []string) error {
	var inBuf = bufio.NewReader(cmd.InOrStdin())

	name := args[0]

	//////////////////////////
	//  Mnemonic Generation //
	//////////////////////////

	// Get bip39 mnemonic
	var mnemonic string
	var bip39Passphrase string = ""

	// User can recover private key from mnemonic
	recover, err := cmd.Flags().GetBool(flagRecover)
	if err != nil {
		return err
	}

	if recover {
		mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	} else {
		// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			return err
		}
	}

	/////////////////////
	// Keys Generation //
	/////////////////////

	var kb *leveldb.DB
	if kb, err = leveldb.OpenFile("/.keybase/poktKeys.db", nil); err != nil {
		return err
	}

	defer kb.Close() // execute at the conclusion of the function

	// Creating a private key with ED25519 and mnemonic seed phrases
	privateKey, err := cryptoPocket.NewPrivateKeyFromSeed([]byte(mnemonic + bip39Passphrase))
	if err != nil {
		return err
	}

	keystore := key{name, privateKey.PublicKey(), privateKey, privateKey.Address(), mnemonic}

	//////////////////
	// Storing keys //
	//////////////////

	// TODO: ask users for passphrase for key protection
	var data []byte
	if data, err = json.Marshal(keystore); err != nil {
		return err
	}

	if err = kb.Put([]byte(name), data, nil); err != nil {
		return err
	}

	///////////////
	// Print Key //
	///////////////
	if err = printKey(keystore); err != nil {
		return err
	}

	return nil
}

func init() {

	// Local Flags
	f := CreateCmd.Flags()
	f.Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
}

// Utility functions

// Print out key in indented JSON format
func printKey(keystore key) error {
	output, err := json.MarshalIndent(keystore, "", "\t")
	if err != nil {
		return err
	}
	log.Printf("%s\n", output)

	return nil
}

// Encryption function that takes a 32 bytes hex string key
func encrypt(stringToEncrypt string, keyString string) (string, error) {
	var err error

	// convert key and plaintext to bytes
	var key, plaintext []byte
	if key, err = hex.DecodeString(keyString); err != nil {
		return "", err
	}
	plaintext = []byte(stringToEncrypt)

	if len(key) != 32 {
		return "", errors.New("key size much be 32 bytes for AES-256 security level")
	}

	// create a new cipher Block from the key
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return "", err
	}

	// create a new GCM (Galois Counter Mode)
	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	var aesGCM cipher.AEAD
	if aesGCM, err = cipher.NewGCM(block); err != nil {
		return "", err
	}

	// create nonce from GCM (nonce size = 12)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// encrypt the data using the AEA GCM Seal function
	enc := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return hex.EncodeToString(enc), nil
}

// Decryption function that takes a 32 bytes hex string key
func decrypt(encryptedString string, keyString string) (string, error) {
	var err error

	// convert key and encrypted data to bytes
	var key, enc []byte
	if key, err = hex.DecodeString(keyString); err != nil {
		return "", err
	}
	if enc, err = hex.DecodeString(encryptedString); err != nil {
		return "", err
	}

	if len(key) != 32 {
		return "", errors.New("key size much be 32 bytes for AES-256 security level")
	}

	// create a new cipher Block from the key
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return "", err
	}

	// create a new GCM (Galois Counter Mode)
	var aesGCM cipher.AEAD
	if aesGCM, err = cipher.NewGCM(block); err != nil {
		return "", err
	}

	// get nonce size and extract nonce from the prefix of the ciphertext
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	// decrypt the data
	var plaintext []byte
	if plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil); err != nil {
		return "", err
	}

	return string(plaintext), nil
}

/* Generating strong random passphrase for user
In most use cases `generateKeyFromPassPhrase` should be sufficient

Warning: this passphrase is not saved by default.
Warning: recommend to store this pass phase in safe vault for users in case they lost it.

*/
func generateRandomKey() (string, error) {
	// generate a random 32 byte key for AES-256
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Keep the key as a secrete! (User only need to remember their passphrase, could be "")
	passphrase := hex.EncodeToString(bytes)
	log.Printf("Generated user passphrase: %s\n", passphrase)
	log.Println("Warning: write this passphrase down!")
	log.Println("Warning: this passphrase will disappear!")
	return passphrase, nil
}

/* Generate a 32 byte hash digest as key based on the user provided pass phrase.
   SHA3-256 only has 128-bit collision resistance, because its output length is 32 bytes.

- Input
	- stringToHash: string of any length

*/
func generateKeyFromPassPhrase(passphrase string) (string, error) {
	// generate a strong random secret key that is at least 32 bytes long
	buf := []byte(passphrase)
	h := cryptoPocket.SHA3Hash(buf)

	// check key length
	if len(h) != 32 {
		return "", errors.New("key generation error: hash digest must be 32 bytes long")
	}

	// return the 32 bytes key hex string
	return hex.EncodeToString(h), nil
}