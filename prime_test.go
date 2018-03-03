package main

import (
    "math"
    "testing"
)

func isPrime(n int32) bool {
    if (n==2)||(n==3) {return true;}
    if n%2 == 0 { return false }
    if n%3 == 0 { return false }
    sqrt := int32(math.Sqrt(float64(n)))
    for i := int32(5); i <= sqrt; i+=6 {
        if n%i == 0 { return false }
        if n%(i+2) == 0 { return false; }
    }
    return true
}

func TestPrime(t *testing.T){
	prime := int32(65533051)
	if !isPrime(prime){
		t.Error(prime, "is actually prime")
	}
}

func BenchmarkPrime(b *testing.B){
	prime := int32(65533051)
	for i :=0; i< b.N; i++ {
		isPrime(prime)
	}
}
