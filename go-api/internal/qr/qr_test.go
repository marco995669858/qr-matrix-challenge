package qr

import (
	"math"
	"testing"
)

const tol = 1e-9

// multiply devuelve a·b.
func multiply(a, b [][]float64) [][]float64 {
	out := make([][]float64, len(a))
	for i := range a {
		out[i] = make([]float64, len(b[0]))
		for j := range b[0] {
			for k := range b {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return out
}

func transpose(a [][]float64) [][]float64 {
	t := make([][]float64, len(a[0]))
	for j := range t {
		t[j] = make([]float64, len(a))
		for i := range a {
			t[j][i] = a[i][j]
		}
	}
	return t
}

func assertClose(t *testing.T, name string, got, want [][]float64) {
	t.Helper()
	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > tol {
				t.Fatalf("%s[%d][%d] = %v, se esperaba %v", name, i, j, got[i][j], want[i][j])
			}
		}
	}
}

// Las tres propiedades que definen una QR correcta.
func TestFactorizeProperties(t *testing.T) {
	cases := map[string][][]float64{
		"cuadrada":         {{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}},
		"alta (m > n)":     {{1, 2}, {3, 4}, {5, 6}},
		"ancha (m < n)":    {{1, 2, 3}, {4, 5, 6}},
		"una fila":         {{3, 4, 5}},
		"una columna":      {{3}, {4}},
		"con ceros":        {{0, 1}, {0, 0}, {1, 0}},
		"ya es triangular": {{2, 1}, {0, 3}},
	}

	for name, a := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := Factorize(a)
			if err != nil {
				t.Fatal(err)
			}
			m := len(a)

			// 1. Q·R reconstruye A.
			assertClose(t, "Q·R", multiply(res.Q, res.R), a)

			// 2. Q es ortogonal: Qᵀ·Q = I.
			assertClose(t, "Qᵀ·Q", multiply(transpose(res.Q), res.Q), identity(m))

			// 3. R es triangular superior.
			for i := range res.R {
				for j := 0; j < i && j < len(res.R[i]); j++ {
					if res.R[i][j] != 0 {
						t.Fatalf("R[%d][%d] = %v, debería ser 0", i, j, res.R[i][j])
					}
				}
			}
		})
	}
}

// Ejemplo clásico de libro (Wikipedia, "QR decomposition").
func TestFactorizeKnownExample(t *testing.T) {
	res, err := Factorize([][]float64{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}})
	if err != nil {
		t.Fatal(err)
	}
	assertClose(t, "R", res.R, [][]float64{
		{14, 21, -14},
		{0, 175, -70},
		{0, 0, -35}, // la QR es única salvo signos; coincide con el ejemplo de referencia
	})
}

func TestFactorizeDoesNotMutateInput(t *testing.T) {
	a := [][]float64{{1, 2}, {3, 4}}
	if _, err := Factorize(a); err != nil {
		t.Fatal(err)
	}
	assertClose(t, "A", a, [][]float64{{1, 2}, {3, 4}})
}

func TestValidate(t *testing.T) {
	cases := map[string][][]float64{
		"vacía":          {},
		"fila vacía":     {{}},
		"no rectangular": {{1, 2}, {3}},
	}
	for name, a := range cases {
		t.Run(name, func(t *testing.T) {
			if err := Validate(a); err == nil {
				t.Fatal("se esperaba un error")
			}
		})
	}

	big := make([][]float64, MaxDimension+1)
	for i := range big {
		big[i] = []float64{1}
	}
	if err := Validate(big); err == nil {
		t.Fatal("se esperaba un error por tamaño")
	}
}
