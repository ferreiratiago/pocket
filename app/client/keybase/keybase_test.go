package keybase

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

//nolint:gosec // G101 Credentials are for tests
const (
	// Example account
	testPrivString    = "045e8380086abc6f6e941d6fe47ca93b86723bc246ec8c4beee411b410028675ed78c49592f836f7a4d47d4fb6a0e6b19f07aebc201d005f6b2c6afe389086e9"
	testPubString     = "ed78c49592f836f7a4d47d4fb6a0e6b19f07aebc201d005f6b2c6afe389086e9"
	testAddr          = "26e16ccab7a898400022476332e2972b8199f2f9"
	testChildAddrIdx1 = "8b83d7057df7ac1d20a2f0aa0edadf206eb6764d"

	// Other
	testPassphrase    = "Testing@Testing123"
	testNewPassphrase = "321gnitsetgnitset"
	testHint          = "testing"
	testTx            = "79fca587bbcfd5da86d73e1d849769017b1c91cc8177dec0fc0e3e0d345f2b35"

	// JSON account
	testJSONAddr       = "572f306e2d29cb8d77c02ebed7d11a5750c815f2"
	testJSONPubString  = "408bec6320b540aa0cc86b3e633e214f2fd4dce4caa08f164fa3a9d3e577b46c"
	testJSONPrivString = "3554119cec1c0c8c5b3845a5d3fc6346eb44ed21aab5c063ae9b6b1d38bec275408bec6320b540aa0cc86b3e633e214f2fd4dce4caa08f164fa3a9d3e577b46c"
	testJSONString     = `{"kdf":"scrypt","salt":"197d2754445a7e5ce3e6c8d7b1d0ff6f","secparam":"12","hint":"pocket wallet","ciphertext":"B/AORJrSeQrR5ewQGel4FeCCXscoCsMUzq9gXAAxDqjXMmMxa7TedBTuemtO82JyTCoQWFHbGxRx8A7IoETNh5T5yBAjNNrr7DDkVrcfSAM3ez9lQem17DsfowCvRtmbesDlvbSZMRy8mQgClLqWRN+c6W/fPQ/lxLUy1G1A965U/uImcMXzSwbfqYrBPEux"}`
)

func TestKeybase_CreateNewKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.Create(testPassphrase, testHint)
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, len(key.GetAddressBytes()), crypto.AddressLen)

	_, err = key.Unarmour(testPassphrase)
	require.NoError(t, err)
}

func TestKeybase_CreateNewKeyNoPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.Create("", "")
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, len(key.GetAddressBytes()), crypto.AddressLen)

	_, err = key.Unarmour("")
	require.NoError(t, err)
}

func TestKeybase_ImportKeyFromString(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)

	require.Equal(t, len(key.GetAddressBytes()), crypto.AddressLen)
	require.Equal(t, key.GetAddressString(), testAddr)
	require.Equal(t, key.GetPublicKey().String(), testPubString)

	privKey, err := key.Unarmour(testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privKey.String(), testPrivString)
}

func TestKeybase_ImportKeyFromStringNoPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.ImportFromJSON(testJSONString, testPassphrase)
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, len(key.GetAddressBytes()), crypto.AddressLen)
	require.Equal(t, key.GetAddressString(), testJSONAddr)
	require.Equal(t, key.GetPublicKey().String(), testJSONPubString)

	privKey, err := keypair.Unarmour(testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privKey.String(), testJSONPrivString)
}

// TODO: Improve this test/create functions to check string validity
func TestKeybase_ImportKeyFromStringInvalidString(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	falseAddr := testKey.String() + "aa"
	falseBz, err := hex.DecodeString(falseAddr)
	require.NoError(t, err)

	keypair, err := db.ImportFromString(falseAddr, testPassphrase, testHint)
	require.EqualError(t, err, crypto.ErrInvalidPrivateKeyLen(len(falseBz)).Error())
	require.Nil(t, keypair)
}

func TestKeybase_ImportKeyFromJSON(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.ImportFromJSON(testJSONString, testPassphrase)
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, len(key.GetAddressBytes()), crypto.AddressLen)
	require.Equal(t, key.GetAddressString(), testJSONAddr)
	require.Equal(t, key.GetPublicKey().String(), testJSONPubString)

	privKey, err := key.Unarmour(testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privKey.String(), testJSONPrivString)
}

func TestKeybase_GetKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	key, err := db.Get(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), key.GetAddressBytes())
	require.Equal(t, key.GetAddressString(), testKey.Address().String())

	privKey, err := keypair.Unarmour(testPassphrase)
	require.NoError(t, err)

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_GetKeyDoesntExist(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.Get(testKey.Address().String())
	require.EqualError(t, err, ErrorAddrNotFound(testKey.Address().String()).Error())
	require.Equal(t, keypair, nil)
}

func TestKeybase_GetAllKeys(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	pkm := make(map[string]crypto.PrivateKey, 0)
	pks := createTestKeys(t, 5)
	for i := 0; i < 5; i++ {
		keypair, err := db.ImportFromString(pks[i].String(), testPassphrase, testHint)
		require.NoError(t, err)
		require.NotNil(t, keypair)
		pkm[pks[i].Address().String()] = pks[i]
	}

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)
	require.Equal(t, len(keypairs), 5)

	for i := 0; i < 5; i++ {
		privKey, err := keypairs[i].Unarmour(testPassphrase)
		require.NoError(t, err)

		require.Equal(t, addresses[i], keypairs[i].GetAddressString())
		require.Equal(t, addresses[i], privKey.Address().String())

		equal := privKey.Equals(pkm[privKey.Address().String()])
		require.Equal(t, equal, true)
	}
}

func TestKeybase_GetPubKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	pubKey, err := db.GetPubKey(keypair.GetAddressString())
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), pubKey.Address().Bytes())
	require.Equal(t, pubKey.Address().String(), testKey.Address().String())

	equal := pubKey.Equals(testKey.PublicKey())
	require.Equal(t, equal, true)
}

func TestKeybase_GetPrivKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(keypair.GetAddressString(), testPassphrase)
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), privKey.Address().Bytes())
	require.Equal(t, privKey.Address().String(), testKey.Address().String())

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_GetPrivKeyWrongPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(keypair.GetAddressString(), testNewPassphrase)
	require.Equal(t, err, crypto.ErrorWrongPassphrase)
	require.Nil(t, privKey)
}

func TestKeybase_UpdatePassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	err = db.UpdatePassphrase(keypair.GetAddressString(), testPassphrase, testNewPassphrase, testHint)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(testKey.Address().String(), testNewPassphrase)
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), privKey.Address().Bytes())
	require.Equal(t, privKey.Address().String(), testKey.Address().String())

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_UpdatePassphraseWrongPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	err = db.UpdatePassphrase(keypair.GetAddressString(), testNewPassphrase, testNewPassphrase, testHint)
	require.ErrorIs(t, err, crypto.ErrorWrongPassphrase)
}

func TestKeybase_DeleteKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	err = db.Delete(keypair.GetAddressString(), testPassphrase)
	require.NoError(t, err)

	delKey, err := db.Get(testKey.Address().String())
	require.EqualError(t, err, ErrorAddrNotFound(testKey.Address().String()).Error())
	require.Equal(t, delKey, nil)
}

func TestKeybase_DeleteKeyWrongPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	testKey := createTestKeys(t, 1)[0]

	keypair, err := db.ImportFromString(testKey.String(), testPassphrase, testHint)
	require.NoError(t, err)

	err = db.Delete(keypair.GetAddressString(), testNewPassphrase)
	require.ErrorIs(t, err, crypto.ErrorWrongPassphrase)
}

func TestKeybase_SignMessage(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	txBz, err := hex.DecodeString(testTx)
	require.NoError(t, err)

	signedMsg, err := db.Sign(keypair.GetAddressString(), testPassphrase, txBz)
	require.NoError(t, err)

	verified, err := db.Verify(keypair.GetAddressString(), txBz, signedMsg)
	require.NoError(t, err)
	require.Equal(t, verified, true)
}

func TestKeybase_SignMessageWrongPassphrase(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	keypair, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	txBz, err := hex.DecodeString(testTx)
	require.NoError(t, err)

	signedMsg, err := db.Sign(keypair.GetAddressString(), testNewPassphrase, txBz)
	require.ErrorIs(t, err, crypto.ErrorWrongPassphrase)
	require.Nil(t, signedMsg)
}

func TestKeybase_ExportString(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	_, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	privStr, err := db.ExportPrivString(testAddr, testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privStr, testPrivString)
}

func TestKeybase_ExportJSON(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	_, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	jsonStr, err := db.ExportPrivJSON(testAddr, testPassphrase)
	require.NoError(t, err)

	err = db.Delete(testAddr, testPassphrase)
	require.NoError(t, err)

	keypair, err := db.ImportFromJSON(jsonStr, testPassphrase)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(testAddr, testPassphrase)
	require.NoError(t, err)
	require.Equal(t, keypair.GetAddressString(), testAddr)
	require.Equal(t, privKey.String(), testPrivString)
}

func TestKeybase_DeriveChildFromKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	_, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	childKey, err := db.DeriveChildFromKey(testAddr, testPassphrase, 1, "", "", false)
	require.NoError(t, err)
	require.Equal(t, childKey.GetAddressString(), testChildAddrIdx1)
}

func TestKeybase_DeriveChildFromSeed(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	kp, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	seed, err := kp.GetSeed(testPassphrase)
	require.NoError(t, err)

	childKey, err := db.DeriveChildFromSeed(seed, 1, "", "", false)
	require.NoError(t, err)
	require.Equal(t, childKey.GetAddressString(), testChildAddrIdx1)
}

func TestKeybase_StoreChildFromKey(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	_, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	childKey, err := db.DeriveChildFromKey(testAddr, testPassphrase, 1, testPassphrase, testHint, true)
	require.NoError(t, err)
	require.NotNil(t, childKey)
	require.Equal(t, childKey.GetAddressString(), testChildAddrIdx1)
}

func TestKeybase_StoreChildFromSeed(t *testing.T) {
	db := initDB(t)
	defer stopDB(t, db)

	kp, err := db.ImportFromString(testPrivString, testPassphrase, testHint)
	require.NoError(t, err)

	seed, err := kp.GetSeed(testPassphrase)
	require.NoError(t, err)

	childKey, err := db.DeriveChildFromSeed(seed, 1, testPassphrase, testHint, true)
	require.NoError(t, err)
	require.NotNil(t, childKey)
	require.Equal(t, childKey.GetAddressString(), testChildAddrIdx1)
}

func initDB(t *testing.T) Keybase {
	db, err := NewKeybaseInMemory()
	require.NoError(t, err)
	return db
}

func createTestKeys(t *testing.T, n int) []crypto.PrivateKey {
	pks := make([]crypto.PrivateKey, 0)
	for i := 0; i < n; i++ {
		privKeyString, _, _ := keygen.GetInstance().Next()
		privKey, err := crypto.NewPrivateKey(privKeyString)
		require.NoError(t, err)
		pks = append(pks, privKey)

	}
	return pks
}

func stopDB(t *testing.T, db Keybase) {
	err := db.Stop()
	require.NoError(t, err)
}
