package main_test

import (
	crand "crypto/rand"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	jnanoid "github.com/jaevor/go-nanoid"
	mnanoid "github.com/matoous/go-nanoid/v2"
	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

func BenchmarkJaevorNanoID(b *testing.B) {
	f, err := jnanoid.Standard(21)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

func BenchmarkUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uuid.NewString()
	}
}

func BenchmarkMatoousNanoID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mnanoid.New()
	}
}

func BenchmarkUlidMake(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.Make()
	}
}

func BenchmarkUlidRand(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.New(ulid.Timestamp(time.Now()), rng)
	}
}

func BenchmarkUlidCryptoRand(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.New(ulid.Timestamp(time.Now()), crand.Reader)
	}
}

func BenchmarkUlidRandMono(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.New(ulid.Timestamp(time.Now()), ulid.Monotonic(rng, 1))
	}
}

func BenchmarkUlidCryptoRandMono(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.New(ulid.Timestamp(time.Now()), ulid.Monotonic(crand.Reader, 1))
	}
}

func BenchmarkKsuidRand(b *testing.B) {
	ksuid.SetRand(ksuid.FastRander)
	for i := 0; i < b.N; i++ {
		ksuid.New()
	}
}

func BenchmarkKsuidCryptoRand(b *testing.B) {
	ksuid.SetRand(nil)
	for i := 0; i < b.N; i++ {
		ksuid.New()
	}
}

func BenchmarkXid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xid.New()
	}
}
