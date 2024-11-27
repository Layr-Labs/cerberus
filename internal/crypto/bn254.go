package crypto

import (
	"encoding/hex"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type G1Point struct {
	*bn254.G1Affine
}

// Add another G1 point to this one
func (p *G1Point) Add(p2 *G1Point) *G1Point {
	p.G1Affine.Add(p.G1Affine, p2.G1Affine)
	return p
}

// Sub another G1 point from this one
func (p *G1Point) Sub(p2 *G1Point) *G1Point {
	p.G1Affine.Sub(p.G1Affine, p2.G1Affine)
	return p
}

// VerifyEquivalence verifies G1Point is equivalent the G2Point
func (p *G1Point) VerifyEquivalence(p2 *G2Point) (bool, error) {
	return CheckG1AndG2DiscreteLogEquality(p.G1Affine, p2.G2Affine)
}

func (p *G1Point) Serialize() []byte {
	return SerializeG1(p.G1Affine)
}

func (p *G1Point) Deserialize(data []byte) *G1Point {
	return &G1Point{DeserializeG1(data)}
}

type G2Point struct {
	*bn254.G2Affine
}

// Add another G2 point to this one
func (p *G2Point) Add(p2 *G2Point) *G2Point {
	p.G2Affine.Add(p.G2Affine, p2.G2Affine)
	return p
}

// Sub another G2 point from this one
func (p *G2Point) Sub(p2 *G2Point) *G2Point {
	p.G2Affine.Sub(p.G2Affine, p2.G2Affine)
	return p
}

func (p *G2Point) Serialize() []byte {
	return SerializeG2(p.G2Affine)
}

func (p *G2Point) Deserialize(data []byte) *G2Point {
	return &G2Point{DeserializeG2(data)}
}

type Signature struct {
	*G1Point `json:"g1_point"`
}

func (s *Signature) Add(otherS *Signature) *Signature {
	s.G1Point.Add(otherS.G1Point)
	return s
}

// Verify a message against a public key
func (s *Signature) Verify(pubkey *G2Point, message [32]byte) (bool, error) {
	ok, err := VerifySig(s.G1Affine, pubkey.G2Affine, message)
	if err != nil {
		return false, err
	}
	return ok, nil
}

type PrivateKey = fr.Element

type KeyPair struct {
	PrivKey *PrivateKey
	PubKey  *G1Point
}

func NewKeyPair(sk *PrivateKey) *KeyPair {
	pk := MulByGeneratorG1(sk)
	return &KeyPair{sk, &G1Point{pk}}
}

// NewKeyPairFromString creates a new keypair from a decimal string
func NewKeyPairFromString(sk string) (*KeyPair, error) {
	ele, err := new(fr.Element).SetString(sk)
	if err != nil {
		return nil, err
	}
	return NewKeyPair(ele), nil
}

// NewKeyPairFromHexString creates a new keypair from a hex string
func NewKeyPairFromHexString(sk string) (*KeyPair, error) {
	skBytes, err := hex.DecodeString(sk)
	if err != nil {
		return nil, err
	}
	skInt := new(big.Int).SetBytes(skBytes)
	ele, err := new(fr.Element).SetString(skInt.String())
	if err != nil {
		return nil, err
	}
	return NewKeyPair(ele), nil
}

// SignMessage This signs a message on G1, and so will require a G2Pubkey to verify
func (k *KeyPair) SignMessage(message [32]byte) *Signature {
	H := MapToCurve(message)
	sig := new(bn254.G1Affine).ScalarMultiplication(H, k.PrivKey.BigInt(new(big.Int)))
	return &Signature{&G1Point{sig}}
}

// SignHashedToCurveMessage This signs a message on G1, and so will require a G2Pubkey to verify
func (k *KeyPair) SignHashedToCurveMessage(g1HashedMsg *bn254.G1Affine) *Signature {
	sig := new(bn254.G1Affine).ScalarMultiplication(g1HashedMsg, k.PrivKey.BigInt(new(big.Int)))
	return &Signature{&G1Point{sig}}
}

func (k *KeyPair) GetPubKeyG2() *G2Point {
	return &G2Point{MulByGeneratorG2(k.PrivKey)}
}

func (k *KeyPair) GetPubKeyG1() *G1Point {
	return k.PubKey
}
