// Package qr implementa la factorización QR mediante rotaciones de Givens.
//
// Dada una matriz A de m×n, encuentra Q (m×m, ortogonal) y R (m×n,
// triangular superior) tales que A = Q·R.
package qr

import (
	"errors"
	"fmt"
	"math"
)

// MaxDimension limita el tamaño de entrada para evitar abusos
// (la QR completa cuesta O(m²·n)).
const MaxDimension = 100

// Result contiene las dos matrices de la factorización A = Q·R.
type Result struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

// Validate comprueba que la matriz sea rectangular, no vacía y de tamaño razonable.
func Validate(a [][]float64) error {
	if len(a) == 0 {
		return errors.New("la matriz no puede estar vacía")
	}
	if len(a) > MaxDimension {
		return fmt.Errorf("la matriz puede tener como máximo %d filas", MaxDimension)
	}
	cols := len(a[0])
	if cols == 0 || cols > MaxDimension {
		return fmt.Errorf("cada fila debe tener entre 1 y %d columnas", MaxDimension)
	}
	for i, row := range a {
		if len(row) != cols {
			return fmt.Errorf("la fila %d tiene %d columnas y se esperaban %d (la matriz debe ser rectangular)", i, len(row), cols)
		}
	}
	return nil
}

// Factorize calcula la factorización QR completa de a.
// No modifica la matriz de entrada.
func Factorize(a [][]float64) (Result, error) {
	if err := Validate(a); err != nil {
		return Result{}, err
	}

	m, n := len(a), len(a[0])
	r := clone(a)
	q := identity(m)

	// Recorremos cada columna j y anulamos los elementos por debajo de la diagonal.
	for j := 0; j < n && j < m-1; j++ {
		for i := j + 1; i < m; i++ {
			if r[i][j] == 0 {
				continue // ya es cero, no hace falta rotar
			}

			// Rotación de Givens G que actúa sobre las filas j e i:
			//   [ c  s] [r[j][j]]   [h]
			//   [-s  c] [r[i][j]] = [0]
			h := math.Hypot(r[j][j], r[i][j]) // evita overflow frente a sqrt(a²+b²)
			c, s := r[j][j]/h, r[i][j]/h

			// R ← G·R (solo cambian las filas j e i, desde la columna j).
			for k := j; k < n; k++ {
				rj, ri := r[j][k], r[i][k]
				r[j][k] = c*rj + s*ri
				r[i][k] = -s*rj + c*ri
			}

			// Q ← Q·Gᵀ (solo cambian las columnas j e i), para mantener A = Q·R.
			for k := 0; k < m; k++ {
				qj, qi := q[k][j], q[k][i]
				q[k][j] = c*qj + s*qi
				q[k][i] = -s*qj + c*qi
			}

			r[i][j] = 0 // cero exacto, sin ruido de redondeo
		}
	}

	cleanNoise(q)
	cleanNoise(r)
	return Result{Q: q, R: r}, nil
}

// noiseThreshold: valores menores a esto se consideran cero de punto flotante.
const noiseThreshold = 1e-12

// cleanNoise convierte residuos como 1e-17 o -0 en un 0 exacto.
func cleanNoise(mat [][]float64) {
	for _, row := range mat {
		for k, v := range row {
			if math.Abs(v) < noiseThreshold {
				row[k] = 0
			}
		}
	}
}

func identity(n int) [][]float64 {
	id := make([][]float64, n)
	for i := range id {
		id[i] = make([]float64, n)
		id[i][i] = 1
	}
	return id
}

func clone(a [][]float64) [][]float64 {
	c := make([][]float64, len(a))
	for i, row := range a {
		c[i] = append([]float64(nil), row...)
	}
	return c
}
