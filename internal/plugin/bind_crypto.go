package plugin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/dop251/goja"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

const (
	cryptoMaxRandomBytes  = 1024 * 1024
	cryptoPBKDF2MinIter   = 1000
	cryptoPBKDF2MaxIter   = 10_000_000
	cryptoPBKDF2MinKeyLen = 16
	cryptoPBKDF2MaxKeyLen = 1024
	cryptoBcryptMinCost   = 4
	cryptoBcryptMaxCost   = 14
	cryptoBcryptDefault   = 10
)

func (rt *JSRuntime) createCryptoModule() *goja.Object {
	obj := rt.vm.NewObject()

	obj.Set("sha256", func(call goja.FunctionCall) goja.Value {
		sum := sha256.Sum256([]byte(call.Argument(0).String()))
		return rt.vm.ToValue(hex.EncodeToString(sum[:]))
	})

	obj.Set("sha1", func(call goja.FunctionCall) goja.Value {
		sum := sha1.Sum([]byte(call.Argument(0).String()))
		return rt.vm.ToValue(hex.EncodeToString(sum[:]))
	})

	obj.Set("md5", func(call goja.FunctionCall) goja.Value {
		sum := md5.Sum([]byte(call.Argument(0).String()))
		return rt.vm.ToValue(hex.EncodeToString(sum[:]))
	})

	obj.Set("hmacSha256", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		data := call.Argument(1).String()
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(data))
		return rt.vm.ToValue(hex.EncodeToString(mac.Sum(nil)))
	})

	obj.Set("pbkdf2", func(call goja.FunctionCall) goja.Value {
		password := call.Argument(0).String()
		salt := call.Argument(1).String()
		iterations := int(call.Argument(2).ToInteger())
		keyLen := int(call.Argument(3).ToInteger())

		if iterations < cryptoPBKDF2MinIter || iterations > cryptoPBKDF2MaxIter {
			panic(rt.vm.ToValue(fmt.Sprintf("pbkdf2: iterations must be between %d and %d", cryptoPBKDF2MinIter, cryptoPBKDF2MaxIter)))
		}
		if keyLen < cryptoPBKDF2MinKeyLen || keyLen > cryptoPBKDF2MaxKeyLen {
			panic(rt.vm.ToValue(fmt.Sprintf("pbkdf2: keyLen must be between %d and %d", cryptoPBKDF2MinKeyLen, cryptoPBKDF2MaxKeyLen)))
		}

		key := pbkdf2.Key([]byte(password), []byte(salt), iterations, keyLen, sha256.New)
		return rt.vm.ToValue(hex.EncodeToString(key))
	})

	obj.Set("bcrypt", func(call goja.FunctionCall) goja.Value {
		password := call.Argument(0).String()
		cost := cryptoBcryptDefault
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			cost = int(call.Argument(1).ToInteger())
		}
		if cost < cryptoBcryptMinCost || cost > cryptoBcryptMaxCost {
			panic(rt.vm.ToValue(fmt.Sprintf("bcrypt: cost must be between %d and %d", cryptoBcryptMinCost, cryptoBcryptMaxCost)))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("bcrypt: %v", err)))
		}
		return rt.vm.ToValue(string(hash))
	})

	obj.Set("bcryptVerify", func(call goja.FunctionCall) goja.Value {
		password := call.Argument(0).String()
		hash := call.Argument(1).String()
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return rt.vm.ToValue(err == nil)
	})

	obj.Set("randomBytes", func(call goja.FunctionCall) goja.Value {
		n := int(call.Argument(0).ToInteger())
		if n <= 0 || n > cryptoMaxRandomBytes {
			panic(rt.vm.ToValue(fmt.Sprintf("randomBytes: size must be between 1 and %d", cryptoMaxRandomBytes)))
		}
		buf := make([]byte, n)
		if _, err := rand.Read(buf); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("randomBytes: %v", err)))
		}
		return rt.vm.ToValue(hex.EncodeToString(buf))
	})

	obj.Set("randomBase64", func(call goja.FunctionCall) goja.Value {
		n := int(call.Argument(0).ToInteger())
		if n <= 0 || n > cryptoMaxRandomBytes {
			panic(rt.vm.ToValue(fmt.Sprintf("randomBase64: size must be between 1 and %d", cryptoMaxRandomBytes)))
		}
		buf := make([]byte, n)
		if _, err := rand.Read(buf); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("randomBase64: %v", err)))
		}
		return rt.vm.ToValue(base64.RawURLEncoding.EncodeToString(buf))
	})

	obj.Set("uuid", func(call goja.FunctionCall) goja.Value {
		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("uuid: %v", err)))
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		return rt.vm.ToValue(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
	})

	obj.Set("base64encode", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		return rt.vm.ToValue(base64.StdEncoding.EncodeToString([]byte(s)))
	})

	obj.Set("base64decode", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("base64decode: %v", err)))
		}
		return rt.vm.ToValue(string(data))
	})

	obj.Set("hexEncode", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		return rt.vm.ToValue(hex.EncodeToString([]byte(s)))
	})

	obj.Set("hexDecode", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		data, err := hex.DecodeString(s)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("hexDecode: %v", err)))
		}
		return rt.vm.ToValue(string(data))
	})

	obj.Set("aesGcmEncrypt", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		plaintext := call.Argument(1).String()

		out, err := aesGcmEncrypt(key, plaintext)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("aesGcmEncrypt: %v", err)))
		}
		return rt.vm.ToValue(out)
	})

	obj.Set("aesGcmDecrypt", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		ciphertext := call.Argument(1).String()

		out, err := aesGcmDecrypt(key, ciphertext)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("aesGcmDecrypt: %v", err)))
		}
		return rt.vm.ToValue(out)
	})

	obj.Set("constantTimeCompare", func(call goja.FunctionCall) goja.Value {
		a := call.Argument(0).String()
		b := call.Argument(1).String()
		eq := subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
		return rt.vm.ToValue(eq)
	})

	return obj
}

// aesGcmEncrypt шифрует строку с ключом произвольной длины (деривация sha256).
// Возвращает base64(nonce || ciphertext || tag).
func aesGcmEncrypt(key, plaintext string) (string, error) {
	derived := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	out := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// aesGcmDecrypt расшифровывает то, что вернул aesGcmEncrypt.
func aesGcmDecrypt(key, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}

	derived := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, sealed := raw[:ns], raw[ns:]

	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt failed: %w", err)
	}
	return string(plain), nil
}
