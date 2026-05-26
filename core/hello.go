// Package core holds shared library code used by the finops CLI (domain logic without local machine I/O).
//
// hello.go exposes a minimal demo API used by "finops demo hello".
package core

import "context"

func Hello(ctx context.Context) (string, error) {
	return "hello", nil
}
