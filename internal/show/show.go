//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package show

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zalando/rds-health/internal/types"
)

//
// The package defines generic primitives to implement formatted output
//

// Generic printer that translates type T instance to sequence of bytes
type Printer[T any] interface {
	Show(T) ([]byte, error)
}

// Lifts a printer function to Pinter interface
type FromShow[T any] func(T) ([]byte, error)

func (f FromShow[T]) Show(x T) ([]byte, error) { return f(x) }

// Prepend prefix
type Prefix[T any] string

func (p Prefix[T]) FMap(f Printer[T]) Printer[T] {
	return FromShow[T](func(x T) ([]byte, error) {
		b := &bytes.Buffer{}

		v, err := f.Show(x)
		if err != nil {
			return nil, err
		}

		if len(v) != 0 {
			if _, err := b.Write([]byte(p)); err != nil {
				return nil, err
			}

			if _, err := b.Write(v); err != nil {
				return nil, err
			}
		}

		return b.Bytes(), nil
	})
}

// Builds printer for type B from printer of type A and contramap B -> A
type ContraMap[A, B any] struct{ T Printer[A] }

func (c ContraMap[A, B]) FMap(f func(B) A) Printer[B] {
	return FromShow[B](func(a B) ([]byte, error) {
		return c.T.Show(f(a))
	})
}

// Build a printer for sequence
type Seq[T any] struct{ T Printer[T] }

func (seq Seq[T]) Show(x []T) ([]byte, error) {
	b := &bytes.Buffer{}

	for _, k := range x {
		v, err := seq.T.Show(k)
		if err != nil {
			return nil, err
		}

		if len(v) != 0 {
			if _, err := b.Write(v); err != nil {
				return nil, err
			}
		}
	}

	return b.Bytes(), nil
}

type UnApply2[T, A, B any] func(T) (A, B)

// Build printer for product type
type Printer2[T, A, B any] struct {
	A Printer[A]
	B Printer[B]
	UnApply2[T, A, B]
}

func (p Printer2[T, A, B]) Show(x T) ([]byte, error) {
	a, b := p.UnApply2(x)

	c := &bytes.Buffer{}

	va, err := p.A.Show(a)
	if err != nil {
		return nil, err
	}

	if len(va) != 0 {
		if _, err := c.Write(va); err != nil {
			return nil, err
		}
	}

	vb, err := p.B.Show(b)
	if err != nil {
		return nil, err
	}

	if len(vb) != 0 {
		if _, err := c.Write(vb); err != nil {
			return nil, err
		}
	}

	return c.Bytes(), nil
}

func Cluster[T, A any](t Printer[T], a Printer[A], f UnApply2[T, []A, []A]) Printer[T] {
	showNodes := Printer2[T, []A, []A]{
		A:        Seq[A]{T: a},
		B:        Seq[A]{T: a},
		UnApply2: f,
	}

	return Printer2[T, T, T]{
		A:        t,
		B:        showNodes,
		UnApply2: func(x T) (T, T) { return x, x },
	}
}

func Region[T, A, B any](a Printer[A], b Printer[B], f UnApply2[T, []A, []B]) Printer[T] {
	return Printer2[T, []A, []B]{
		A:        Prefix[[]A]("").FMap(Seq[A]{T: a}),
		B:        Prefix[[]B]("\n").FMap(Seq[B]{T: b}),
		UnApply2: f,
	}
}

// outputs json
func JSON[T any]() Printer[T] {
	return FromShow[T](func(x T) ([]byte, error) {
		return json.MarshalIndent(x, "", "  ")
	})
}

// outputs nothing
func None[T any]() Printer[T] {
	return FromShow[T](func(x T) ([]byte, error) {
		return nil, nil
	})
}

type SchemaStatusCode struct {
	NONE string
	PASS string
	WARN string
	FAIL string
}

type Schema struct {
	StatusCodeIcon SchemaStatusCode
	StatusCodeText SchemaStatusCode
	Cluster        string
}

func (s Schema) FmtForStatus(c types.StatusCode) string {
	switch c {
	case types.STATUS_CODE_UNKNOWN:
		return s.StatusCodeText.NONE
	case types.STATUS_CODE_SUCCESS:
		return s.StatusCodeText.PASS
	case types.STATUS_CODE_WARNING:
		return s.StatusCodeText.WARN
	case types.STATUS_CODE_FAILURE:
		return s.StatusCodeText.FAIL
	default:
		return "%s"
	}
}

func StatusText(x types.StatusCode) string {
	switch x {
	case types.STATUS_CODE_UNKNOWN:
		return fmt.Sprintf(SCHEMA.StatusCodeText.NONE, "NONE")
	case types.STATUS_CODE_SUCCESS:
		return fmt.Sprintf(SCHEMA.StatusCodeText.PASS, "PASS")
	case types.STATUS_CODE_WARNING:
		return fmt.Sprintf(SCHEMA.StatusCodeText.WARN, "WARN")
	case types.STATUS_CODE_FAILURE:
		return fmt.Sprintf(SCHEMA.StatusCodeText.FAIL, "FAIL")
	default:
		return ""
	}
}

func StatusIcon(x types.StatusCode) string {
	switch x {
	case types.STATUS_CODE_UNKNOWN:
		return SCHEMA.StatusCodeIcon.NONE
	case types.STATUS_CODE_SUCCESS:
		return SCHEMA.StatusCodeIcon.PASS
	case types.STATUS_CODE_WARNING:
		return SCHEMA.StatusCodeIcon.WARN
	case types.STATUS_CODE_FAILURE:
		return SCHEMA.StatusCodeIcon.FAIL
	default:
		return ""
	}
}

var (
	SCHEMA_PLAIN = Schema{
		StatusCodeIcon: SchemaStatusCode{
			NONE: "",
			PASS: "",
			WARN: "",
			FAIL: "",
		},
		StatusCodeText: SchemaStatusCode{
			NONE: "%s",
			PASS: "%s",
			WARN: "%s",
			FAIL: "%s",
		},

		Cluster: "%s",
	}

	SCHEMA_COLOR = Schema{
		StatusCodeIcon: SchemaStatusCode{
			NONE: "",
			PASS: "✅ ",
			WARN: "🟧 ",
			FAIL: "❌ ",
		},
		StatusCodeText: SchemaStatusCode{
			NONE: "%s",
			PASS: "\033[32m%s\033[0m",
			WARN: "\033[33m%s\033[0m",
			FAIL: "\033[31m%s\033[0m",
		},
		Cluster: "\u001b[1m%s\u001b[0m",
	}

	SCHEMA = SCHEMA_PLAIN
)
