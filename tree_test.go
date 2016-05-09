package tree

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestNode(t *testing.T) {
	tests := []struct {
		path []string
		key  string
		v    interface{}
		err  bool
	}{{
		path: []string{"foo", "bar"},
		key:  "attrib",
		v:    42,
		err:  false,
	}, {
		path: []string{"foo", "bar"},
		key:  "attrib",
		v:    "42",
		err:  false,
	}}
	for _, tt := range tests {
		n := NewNode(tt.path, tt.key, tt.v)
		if tt.err && n != nil || tt.err && n == nil {
			t.Errorf("%v: error check failed: got %v", tt, n)
		}
		if n.Get() != tt.v {
			t.Errorf("%v: Value() failed: got %v, want %v", tt, n.Get(), tt.v)
		}
		if !reflect.DeepEqual(n.path[0:len(n.path)-1], tt.path) {
			t.Errorf("%v: path: got %v, want %v", tt, n.path[0:len(n.path)-1], tt.path)
		}
		if n.path[len(n.path)-1] != tt.key {
			t.Errorf("%v: path: got %v, want %v", tt, n.path[len(n.path)-1], tt.key)
		}
	}
}

func TestTree(t *testing.T) {
	r := New(100)
	if r == nil {
		t.Errorf("New() failed: nil")
	}
	r.Update([]string{"foo"}, 42)
	n, err := r.Get([]string{"foo"})
	if err != nil {
		t.Errorf("Update() failed: error %v", err)
	}
	if got, want := n.Get(), 42; got != want {
		t.Errorf("Update() failed: got %v, want %v", got, want)
	}
	r.Update([]string{"foo", "bar"}, 42)
	r.Update([]string{"seperate", "test", "path"}, 42)
	n = r.Delete([]string{"foo"})
	if n.Get() != 42 {
		t.Errorf("Delete() failed: got %v", n.Get())
	}
	if _, ok := r.edges["foo"]; ok {
		t.Errorf("Delete() failed: edges not removed")
	}
	if _, ok := r.index["foo"]; ok {
		t.Errorf("Delete() failed: index not removed")
	}
}

func BenchmarkLoad(b *testing.B) {
	for l := 0; l < b.N; l++ {
		r := New(100)
		for i := 0; i < 100; i++ {
			for n := 0; n < 10; n++ {
				r.Update([]string{string(n)}, n)
			}
		}
	}
}

func BenchmarkAppend(b *testing.B) {
	maxDepth := 4
	maxChoices := 4
	genPath := func() []string {
		c := []string{}
		for i := 0; i < rand.Intn(maxDepth); i++ {
			c = append(c, string(rand.Intn(maxChoices)))
		}
		return c
	}

	r := New(100)
	for i := 0; i < 100; i++ {
		for n := 0; n < 10; n++ {
			r.Update([]string{string(n)}, n)
		}
	}
	pickPath := func() func() []string {
		p := [][]string{}
		i := 0
		for j := 0; j < 1000000; j++ {
			p = append(p, genPath())
		}
		return func() []string {
			i++
			if i >= len(p) {
				i = 0
			}
			return p[i]
		}
	}
	b.ResetTimer()
	for l := 0; l < b.N; l++ {
		r.Update(pickPath()(), "new")
	}
}
