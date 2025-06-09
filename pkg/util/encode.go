package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func doubleSHA256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

func Base58Encode(input []byte) string {
	var result []byte

	x := new(big.Int).SetBytes(input)

	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	for _, b := range input {
		if b == 0x00 {
			result = append([]byte{base58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return string(result)
}

func Base58CheckEncode(payload []byte) string {
	checksum := doubleSHA256(payload)[:4]

	full := append(payload, checksum...)

	return Base58Encode(full)
}

func base58Decode(input string) ([]byte, error) {
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, r := range input {
		index := strings.IndexRune(base58Alphabet, r)
		if index < 0 {
			return nil, fmt.Errorf("invalid base58 character: %q", r)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(index)))
	}

	decoded := result.Bytes()

	numLeadingZeros := 0
	for _, c := range input {
		if c == rune(base58Alphabet[0]) {
			numLeadingZeros++
		} else {
			break
		}
	}
	decoded = append(bytes.Repeat([]byte{0x00}, numLeadingZeros), decoded...)

	return decoded, nil
}

func Base58CheckDecode(input string) ([]byte, error) {
	full, err := base58Decode(input)
	if err != nil {
		return nil, err
	}

	payload := full[:len(full)-4]
	checksum := full[len(full)-4:]
	expectedChecksum := doubleSHA256(payload)[:4]

	if !bytes.Equal(checksum, expectedChecksum) {
		return nil, errors.New("checksum mismatch")
	}

	return payload, nil
}

func DecodePrivateKey(encodedPrivateKey string) (*ecdsa.PrivateKey, error) {
	privKeyBytes, err := Base58CheckDecode(encodedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %v", err)
	}

	privKeyInt := new(big.Int).SetBytes(privKeyBytes)

	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(privKeyBytes)

	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: privKeyInt,
	}

	return privateKey, nil
}

func EncodePublicKey(privKey *ecdsa.PrivateKey) string {
	publicKeyBytes := elliptic.Marshal(privKey.PublicKey.Curve, privKey.PublicKey.X, privKey.PublicKey.Y)
	publicKeyEncode := Base58CheckEncode(publicKeyBytes)

	return publicKeyEncode
}

func DecodePublicKey(pubKeyEncode string) (*ecdsa.PublicKey, error) {
	decoded, err := Base58CheckDecode(pubKeyEncode)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(elliptic.P256(), decoded)

	return &ecdsa.PublicKey{
		X:     x,
		Y:     y,
		Curve: elliptic.P256(),
	}, nil
}
