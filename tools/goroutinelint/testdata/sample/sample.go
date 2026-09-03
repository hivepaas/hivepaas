// Package sample is fixture code for the goroutinelint tests. It is not compiled
// as part of the app (testdata is ignored by the go tool).
package sample

import (
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/services/docker"
)

func bad1() {
	go func() { // want: unguarded
		println("boom")
	}()
}

func bad2() {
	var wg sync.WaitGroup
	wg.Go(func() { // want: unguarded WaitGroup.Go
		println("boom")
	})
}

func bad3() {
	go unguardedEntryPoint() // want: unguarded callee
}

func unguardedEntryPoint() {
	println("boom")
}

func bad4() {
	go docker.SomeExportedFunc() // want: cannot verify, other package
}

// bad5 has a defer, but it does not recover.
func bad5() {
	go func() { // want: unguarded
		defer println("cleanup")
		println("boom")
	}()
}

// bad6 is the classic trap: recover nested one level too deep never fires.
func bad6() {
	go func() { // want: unguarded
		defer func() { safego.Recover("x") }()
		println("boom")
	}()
}

func ok1() {
	safego.Go("ok1", func() {
		println("safe")
	})
}

func ok2() {
	go func() {
		defer safego.Recover("ok2")
		println("safe")
	}()
}

func ok3() {
	var wg sync.WaitGroup
	wg.Go(func() {
		defer safego.RecoverWithLogger(nil, "ok3")
		println("safe")
	})
}

func ok4() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("safe")
			}
		}()
		println("safe")
	}()
}

func ok5() {
	go guardedEntryPoint()
}

func guardedEntryPoint() {
	defer safego.Recover("ok5")
	println("safe")
}

func ok6() {
	var err error
	go func() {
		defer safego.RecoverTo(&err)
		println("safe")
	}()
}

// bad7 passes nil: the panic has nowhere to go and is silently hidden.
func bad7() {
	go func() { // want: RecoverTo(nil)
		defer safego.RecoverTo(nil)
		println("boom")
	}()
}

func ok7() {
	//safego:allow fixture for the escape hatch
	go func() {
		println("intentionally unguarded")
	}()
}
