package cryptoutil

import "testing"

var benchSecret = []byte("app-secret")

// BenchmarkMakeKey measures one raw argon2id derivation - what every decrypt cost
// before the cache.
func BenchmarkMakeKey(b *testing.B) {
	salt := []byte("some-salt")
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		_ = makeKey(benchSecret, salt)
	}
}

// BenchmarkMakeKeyCachedHit measures the steady state: the same stored secret
// decrypted over and over, which is what a repo webhook does per delivery.
func BenchmarkMakeKeyCachedHit(b *testing.B) {
	salt := []byte("some-salt")
	_ = makeKeyCached(benchSecret, salt) // warm

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = makeKeyCached(benchSecret, salt)
	}
}

// BenchmarkMakeKeyCachedHitParallel checks the RWMutex does not serialize readers
// under the concurrency an unauthenticated webhook route can attract.
func BenchmarkMakeKeyCachedHitParallel(b *testing.B) {
	salt := []byte("some-salt")
	_ = makeKeyCached(benchSecret, salt) // warm

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = makeKeyCached(benchSecret, salt)
		}
	})
}
