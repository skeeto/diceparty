package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/bits"
	"math/rand"
)

var table = [...]int{
	-4, -3, -3, -3, -3, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, +0, +0, +0, +0, +0,
	+0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +1, +1, +1, +1,
	+1, +1, +1, +1, +1, +1, +1, +1, +1, +1, +1, +1, +2, +2, +2, +2, +2, +2,
	+2, +2, +2, +2, +3, +3, +3, +3, +4,
}

type dice struct{ *rand.Rand }

func newDice() *dice {
	var buf [16]byte
	if _, err := crand.Read(buf[:]); err != nil {
		panic(err)
	}
	pcg := pcg64{
		Hi: binary.LittleEndian.Uint64(buf[0:]),
		Lo: binary.LittleEndian.Uint64(buf[8:]),
	}
	return &dice{rand.New(&pcg)}
}

func (d *dice) Roll() int {
	return table[d.Intn(len(table))]
}

// A pcg64 provides a 64-bit permuted congruential generator that
// implements math/rand.Source64. Can be seeded to any value.
type pcg64 struct{ Hi, Lo uint64 }

var _ rand.Source64 = (*pcg64)(nil)

func (s *pcg64) Seed(seed int64) {
	s.Lo = uint64(seed)
	s.Hi = 0
}

func (s *pcg64) Uint64() uint64 {
	const (
		mhi = 0x2360ed051fc65da4
		mlo = 0x4385df649fccf645
		ahi = 0x5851f42d4c957f2d
		alo = 0x14057b7ef767814f
	)
	carry, lo := bits.Mul64(mlo, s.Lo)
	hi := mhi*s.Lo + s.Hi*mlo + carry
	lo, carry = bits.Add64(lo, alo, 0)
	hi += ahi + carry
	s.Lo = lo
	s.Hi = hi
	lo, hi = lo^lo>>43^hi<<21, hi^hi>>43
	r := int(hi>>60) + 45
	return lo>>r | hi<<(64-r)
}

func (s *pcg64) Int63() int64 {
	return int64(s.Uint64() >> 1)
}
