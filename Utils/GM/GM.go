package GM

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
	"github.com/tjfoc/gmsm/x509"
	"io"
	"log"
	"os"
)

var (
	PrivateKey *sm2.PrivateKey
	PublicKey  *sm2.PublicKey
	Header     = "elst.dev"
	DefaultKey = []byte("a2IS83Elst01839S")
	DefaultIV  = []byte("13fd53e24a2779b4")
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// SM4Encrypt SM4加密
func SM4Encrypt(origData, key, IV []byte) ([]byte, error) {
	if key == nil {
		key = DefaultKey
	}
	if IV == nil {
		IV = DefaultIV
	}
	block, err := sm4.NewCipher(key)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, IV)
	cryted := make([]byte, len(origData))
	blockMode.CryptBlocks(cryted, origData)
	result := base64.StdEncoding.EncodeToString(cryted)
	return []byte(result), nil
}

// SM4Decrypt SM4解密
func SM4Decrypt(cryted, key, IV []byte) ([]byte, error) {
	if key == nil {
		key = DefaultKey
	}
	if IV == nil {
		IV = DefaultIV
	}
	cryted, _ = base64.StdEncoding.DecodeString(string(cryted))
	block, err := sm4.NewCipher(key)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, IV)
	origData := make([]byte, len(cryted))
	blockMode.CryptBlocks(origData, cryted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// InitSM2Key 初始化SM2密钥
func InitSM2Key(privateKeyPath, publicKeyPath string) error {
	_, statErr := os.Stat(privateKeyPath)
	file, err := os.OpenFile(privateKeyPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}

	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if publicKeyFile != nil {
		defer publicKeyFile.Close()
	}
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}

	if statErr != nil {
		PrivateKey, err = sm2.GenerateKey(rand.Reader)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			log.Printf("创建密钥异常：%s", err)
			return err
		}
		PublicKey = &PrivateKey.PublicKey
	} else {
		privateByte, err := io.ReadAll(file)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			return err
		}
		privateBlock, _ := pem.Decode(privateByte)
		PrivateKey, err = x509.ReadPrivateKeyFromPem(privateBlock.Bytes, []byte(Header))
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			return err
		}

		publicByte, err := io.ReadAll(publicKeyFile)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			return err
		}
		publicBlock, _ := pem.Decode(publicByte)
		PublicKey, err = x509.ReadPublicKeyFromPem(publicBlock.Bytes)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			return err
		}
		return nil
	}

	sm2PrivateKey, err := x509.WritePrivateKeyToPem(PrivateKey, []byte(Header))
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}

	block := pem.Block{
		Type:    "ELST PRIVATE KEY",
		Headers: nil,
		Bytes:   sm2PrivateKey,
	}
	err = pem.Encode(file, &block)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}

	sm2PublicKey, err := x509.WritePublicKeyToPem(PublicKey)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}
	block2 := pem.Block{
		Type:    "ELST PUBLIC KEY",
		Headers: nil,
		Bytes:   sm2PublicKey,
	}
	err = pem.Encode(publicKeyFile, &block2)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return err
	}
	return nil
}

// SM2PublicEncrypt 公钥加密数据
func SM2PublicEncrypt(origData []byte) ([]byte, error) {
	asn1, err := PublicKey.EncryptAsn1(origData, rand.Reader)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	result := base64.StdEncoding.EncodeToString(asn1)
	return []byte(result), nil
}

// SM2PrivateEncrypt 私钥加密数据
func SM2PrivateEncrypt(origData []byte) ([]byte, error) {
	asn1, err := PrivateKey.EncryptAsn1(origData, rand.Reader)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	result := base64.StdEncoding.EncodeToString(asn1)
	return []byte(result), nil
}

// SM2PublicDecrypt 公钥解密数据
func SM2PublicDecrypt(ciphertext []byte) ([]byte, error) {
	ciphertext, _ = base64.StdEncoding.DecodeString(string(ciphertext))
	origData, err := sm2.DecryptAsn1(PrivateKey, ciphertext)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	return origData, nil
}

// SM2PrivateDecrypt 私钥解密并生成签名
func SM2PrivateDecrypt(ciphertext []byte) (sign []byte, origData []byte, err error) {
	ciphertext, _ = base64.StdEncoding.DecodeString(string(ciphertext))
	origData, err = PrivateKey.DecryptAsn1(ciphertext)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, nil, err
	}
	sign, err = PrivateKey.Sign(rand.Reader, origData, nil)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, nil, err
	}
	return sign, origData, nil
}

// SM2PublicVerifySign 签名确认
func SM2PublicVerifySign(origData, sign []byte) bool {
	verify := PublicKey.Verify(origData, sign)
	return verify
}

func SM3SUM(in string) string {
	sm3Sum := sm3.Sm3Sum([]byte(in))
	return hex.EncodeToString(sm3Sum)
}
