package settings

import (
	"fmt"
	"sync"
	"testing"
)

type stubDB struct{}

func (s *stubDB) CloseDB()                             {}
func (s *stubDB) Get(xPath, name string) []byte        { return []byte("v") }
func (s *stubDB) Set(xPath, name string, value []byte) {}
func (s *stubDB) List(xPath string) []string           { return []string{"a", "b"} }
func (s *stubDB) Rem(xPath, name string)               {}
func (s *stubDB) Clear(xPath string)                   {}

func TestDBReadCacheConcurrentSetList(t *testing.T) {
	cdb := NewDBReadCache(&stubDB{})

	var wg sync.WaitGroup
	const iters = 2000

	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			cdb.Set("viewed", fmt.Sprintf("n%d", i), []byte("x"))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			cdb.Rem("viewed", fmt.Sprintf("n%d", i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			cdb.List("viewed")
		}
	}()
	wg.Wait()
}
