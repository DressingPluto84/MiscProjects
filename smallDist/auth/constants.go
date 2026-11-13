package main

import (
	"crypto/rsa"
	"math/big"
)

type bigIntError struct {
	message string
}

type cryptNums struct {
	mod             *big.Int
	publicExponent  int
	privateExponent *big.Int
	prime1          *big.Int
	prime2          *big.Int
	exponent1       *big.Int
	exponent2       *big.Int
	coefficient     *big.Int
}

func (e *bigIntError) Error() string {
	return e.message
}

func getNums() *cryptNums {
	var mod = new(big.Int)
	_, stat := mod.SetString("", 10)
	if !stat {
		return nil
	}

	var publicExponent = 65537

	var privateExponent = new(big.Int)
	_, stat = privateExponent.SetString("", 10)
	if !stat {
		return nil
	}

	var prime1 = new(big.Int)
	_, stat = prime1.SetString("", 10)
	if !stat {
		return nil
	}

	var prime2 = new(big.Int)
	_, stat = prime2.SetString("", 10)
	if !stat {
		return nil
	}

	var exponent1 = new(big.Int)
	_, stat = exponent1.SetString("", 10)
	if !stat {
		return nil
	}

	var exponent2 = new(big.Int)
	_, stat = exponent2.SetString("", 10)
	if !stat {
		return nil
	}

	var coefficient = new(big.Int)
	_, stat = coefficient.SetString("", 10)
	if !stat {
		return nil
	}

	return &cryptNums{mod, publicExponent, privateExponent, prime1, prime2, exponent1, exponent2, coefficient}
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	nums := getNums()
	if nums == nil {
		return nil, &bigIntError{"Failed To Get Nums"}
	}

	pubKey := rsa.PublicKey{N: nums.mod, E: nums.publicExponent}

	primes := []*big.Int{nums.prime1, nums.prime2}
	precomputed := rsa.PrecomputedValues{Dp: nums.exponent1, Dq: nums.exponent2, Qinv: nums.coefficient}
	privKey := rsa.PrivateKey{PublicKey: pubKey, D: nums.privateExponent, Primes: primes, Precomputed: precomputed}

	return &privKey, nil
}

func getPublicKey() (*rsa.PublicKey, error) {
	nums := getNums()
	if nums == nil {
		return nil, &bigIntError{"Failed To Get Nums"}
	}

	pubKey := rsa.PublicKey{N: nums.mod, E: nums.publicExponent}
	return &pubKey, nil
}
