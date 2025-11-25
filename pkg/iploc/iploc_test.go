package iploc

import (
	"math/rand"
	"testing"
	"time"
)

const DATPath = "/var/teamsacs/data/qqwry.dat"

func TestFind(t *testing.T) {
	loc, err := Open(DATPath)
	if err != nil {
		t.Fatal(err)
	}

	detail1 := loc.Find("4.4")
	detail2, _ := Find(DATPath, "4.0.0.4")
	t.Log(detail1.Location.String(), detail2.Location.String())
	t.Log(detail1.GetCountry())
	t.Log(detail1.GetCounty())
	t.Log(detail1.GetCity())
	t.Log(detail1.GetRegion())
	t.Log(detail1.GetProvince())

	if detail1.Start.Compare(detail2.Start) != 0 || detail1.End.Compare(detail2.End) != 0 || detail1.String() != detail2.String() {
		t.Fatal("error")
	}
}

func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	loc, err := Open(DATPath)
	if err != nil {
		b.Fatal(err)
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		loc.FindUint(rand.Uint32())
	}
}

func BenchmarkFindUnIndexed(b *testing.B) {
	b.StopTimer()
	loc, err := OpenWithoutIndexes(DATPath)
	if err != nil {
		b.Fatal(err)
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		loc.FindUint(rand.Uint32())
	}
}

func BenchmarkFindParallel(b *testing.B) {
	b.StopTimer()
	loc, err := Open(DATPath)
	if err != nil {
		b.Fatal(err)
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if loc.FindUint(rand.Uint32()) == nil {
				b.Fatal("error")
			}
		}
	})
}
