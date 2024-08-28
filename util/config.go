package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/scrypt"
)

type DBEntry struct {
	Name    string `toml:"name"`
	Driver  string `toml:"driver"`
	ConnStr string `toml:"conn_str"`
}

type SqlineConf struct {
	SavedConns []DBEntry `toml:"saved_conns"`
}

func SaveConf(conf *SqlineConf) error {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	confDir = filepath.Join(confDir, "sqline")
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		os.Mkdir(confDir, os.ModeDir)
	}

	f, err := os.OpenFile(filepath.Join(confDir, "conf.toml"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)

	err = encoder.Encode(conf)

	return err
}

func LoadConf() (*SqlineConf, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return &SqlineConf{}, err
	}

	confDir = filepath.Join(confDir, "sqline")
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		return &SqlineConf{}, nil
	}

	confFile, err := os.ReadFile(filepath.Join(confDir, "conf.toml"))
	if err != nil {
		return &SqlineConf{}, err
	}

	var conf SqlineConf
	_, err = toml.Decode(string(confFile), &conf)
	if err != nil {
		return &SqlineConf{}, err
	}

	return &conf, nil
}

func encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nonce, nonce, data, nil)

	cipherText = append(cipherText, salt...)

	return cipherText, nil
}

func decrypt(key, data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce, cipherText := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}
