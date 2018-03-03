package main

import (
    "math"
    "testing"
	"math/big"
)

func i32Prime(n int32) bool {
//    if (n==2)||(n==3) {return true;}
    if n%2 == 0 { return false }
    if n%3 == 0 { return false }
    sqrt := int32(math.Sqrt(float64(n)))
    for i := int32(5); i <= sqrt; i+=6 {
        if n%i == 0 { return false }
        if n%(i+2) == 0 { return false; }
    }
    return true
}

//const num = 65533051
const num = 1 << 19 -1
//const num = 2017133

func TestPrime(t *testing.T){
	prime := int32(num)
	if !i32Prime(prime){
		t.Error(prime, "is actually prime")
	}
}

func BenchmarkPrime(b *testing.B){
	prime := int32(num)
	for n :=0; n < b.N; n++ {
		i32Prime(prime)
	}
}

func BenchmarkPPrime(b *testing.B){
	prime := big.NewInt(num)
	for n :=0; n < b.N; n++ {
		prime.ProbablyPrime(0)
	}
}
