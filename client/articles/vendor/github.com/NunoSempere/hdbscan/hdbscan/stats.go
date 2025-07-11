// Copyright 2020 Humility AI Incorporated, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hdbscan

import (
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// GeneralizedVariance will return the determinant of the covariance matrix
// of the supplied data.
// The supplied data is a list of 'rows' observations of length 'columns'.
func GeneralizedVariance(rows, columns int, data []float64) float64 {
	covMatrix := &mat.SymDense{}
	matrix := mat.NewDense(rows, columns, data)
	stat.CovarianceMatrix(covMatrix, matrix, nil)
	det, _ := mat.LogDet(covMatrix)
	return math.Abs(det)
}
