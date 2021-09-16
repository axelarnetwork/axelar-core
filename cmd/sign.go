package main

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
)

func main() {
	privKeyBz1, _ := hex.DecodeString("8c11063a369d036348c0244c1b7df5eac97d49e0c48c106297a9848f2fe98031")
	privKeyBz2, _ := hex.DecodeString("dde55fe086a1179c5f74f192f7ce2f494ffc70cf59b67daf9255dd9b333cb1ff")
	privKeyBz3, _ := hex.DecodeString("ac2a5e6b9846296da1bb54005be0b4613f1610d6b5b9888a012f0563e8da00e2")
	privKeyBz4, _ := hex.DecodeString("a132e1ad43c7d3f6680a986a00ff76931b17db2694def23733891940cfd75cf4")
	privKeyBz5, _ := hex.DecodeString("86b870b78247c1e29396357eec5a9f55a01e0de2f43cf93e94ff0c9e1b5c974c")
	privKeyBz6, _ := hex.DecodeString("4362cc2156e90723d3ea6f09c04226a39329463e8a0fc5425e076bd7d228f9e1")

	privKey1, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz1)
	privKey2, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz2)
	privKey3, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz3)
	privKey4, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz4)
	privKey5, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz5)
	privKey6, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz6)

	pubKey1 := btcec.PublicKey(privKey1.PublicKey)
	pubKey2 := btcec.PublicKey(privKey2.PublicKey)
	pubKey3 := btcec.PublicKey(privKey3.PublicKey)
	pubKey4 := btcec.PublicKey(privKey4.PublicKey)
	pubKey5 := btcec.PublicKey(privKey5.PublicKey)
	pubKey6 := btcec.PublicKey(privKey6.PublicKey)

	fmt.Printf("pubKey1 %s\n", hex.EncodeToString(pubKey1.SerializeCompressed()))
	fmt.Printf("pubKey2 %s\n", hex.EncodeToString(pubKey2.SerializeCompressed()))
	fmt.Printf("pubKey3 %s\n", hex.EncodeToString(pubKey3.SerializeCompressed()))
	fmt.Printf("pubKey4 %s\n", hex.EncodeToString(pubKey4.SerializeCompressed()))
	fmt.Printf("pubKey5 %s\n", hex.EncodeToString(pubKey5.SerializeCompressed()))
	fmt.Printf("pubKey6 %s\n", hex.EncodeToString(pubKey6.SerializeCompressed()))

	digest, _ := hex.DecodeString("234f6a187520f67969766bdea412e0efaf01e80161b550a6dea9188b7a29be01")

	sig1, _ := privKey1.Sign(digest)
	sig2, _ := privKey2.Sign(digest)
	sig3, _ := privKey3.Sign(digest)
	sig4, _ := privKey4.Sign(digest)
	sig5, _ := privKey5.Sign(digest)
	sig6, _ := privKey6.Sign(digest)

	fmt.Printf("sig1 %s\n", hex.EncodeToString(sig1.Serialize()))
	fmt.Printf("sig2 %s\n", hex.EncodeToString(sig2.Serialize()))
	fmt.Printf("sig3 %s\n", hex.EncodeToString(sig3.Serialize()))
	fmt.Printf("sig4 %s\n", hex.EncodeToString(sig4.Serialize()))
	fmt.Printf("sig5 %s\n", hex.EncodeToString(sig5.Serialize()))
	fmt.Printf("sig6 %s\n", hex.EncodeToString(sig6.Serialize()))
}
