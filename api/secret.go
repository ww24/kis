package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
)

// generate or load auth secret
func authSecret(filepath string) (keyData JSON) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	keyData = JSON{}
	if len(data) > 2 {
		err = json.Unmarshal(data, &keyData)
		if err != nil {
			panic(err)
		}
	}
	if _, ok := keyData["secret"]; ok == false {
		randomBytes := make([]byte, 0x16)
		_, err = rand.Read(randomBytes)
		if err != nil {
			panic(err)
		}
		keyData["secret"] = hex.EncodeToString(randomBytes)

		var data []byte
		data, err = json.Marshal(keyData)
		if err != nil {
			panic(err)
		}
		_, err = file.WriteAt(data, 0)
		if err != nil {
			panic(err)
		}
	}

	return
}

func hashCompare(userSecret string, serverSecret string) (res bool) {
	// userSecret or serverSecret is not set
	if len(userSecret) == 0 || len(serverSecret) == 0 {
		return
	}

	data := sha256.Sum256([]byte(userSecret))
	userSecretHash := hex.EncodeToString(data[:])
	res = userSecretHash == hashedSecret

	// serverSecret is not pre hashed
	if res == false {
		data := sha256.Sum256([]byte(serverSecret))
		serverSecretHash := hex.EncodeToString(data[:])
		res = userSecretHash == serverSecretHash
	}
	return
}
