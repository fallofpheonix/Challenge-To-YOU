package simulation

import "math/rand"

// DetRNG wraps math/rand.Rand and tracks the total number of Int63 calls made.
// The (seed, callCount) pair uniquely identifies any position in the random sequence,
// so Engine.SetState can reconstruct the exact RNG state by fast-forwarding from the seed.
//
// Why call-count tracking instead of exposing rand.Rand internals:
//   math/rand.Rand does not export its source state, making serialization impossible
//   without wrapping. Tracking calls is O(1) overhead per call; restoration is O(callCount),
//   which for 500-tick checkpoints at ~5 calls/drone/tick is at most ~250,000 iterations —
//   well under 1 ms.
type DetRNG struct {
	seed      int64
	callCount int64
	inner     *rand.Rand
}

func newDetRNG(seed int64) *DetRNG {
	return &DetRNG{
		seed:  seed,
		inner: rand.New(rand.NewSource(seed)),
	}
}

func (r *DetRNG) Intn(n int) int {
	r.callCount++
	return r.inner.Intn(n)
}

// snapshot returns the (seed, callCount) pair. Store both in a checkpoint to
// reconstruct this exact RNG position later without replaying the full simulation.
func (r *DetRNG) snapshot() (seed, callCount int64) {
	return r.seed, r.callCount
}

// restoreDetRNG recreates a DetRNG at the given (seed, callCount) position by
// replaying callCount Int63 draws from the seed. This is the inverse of snapshot().
func restoreDetRNG(seed, callCount int64) *DetRNG {
	r := &DetRNG{
		seed:  seed,
		inner: rand.New(rand.NewSource(seed)),
	}
	for i := int64(0); i < callCount; i++ {
		r.inner.Int63()
	}
	r.callCount = callCount
	return r
}
