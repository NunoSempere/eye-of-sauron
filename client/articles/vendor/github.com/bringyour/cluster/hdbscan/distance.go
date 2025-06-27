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
)

// DistanceFunc ...
type DistanceFunc func(x1, x2 []float64) float64

// EuclideanDistance ...
var EuclideanDistance = func(v1, v2 []float64) float64 {
	acc := 0.0
	for i, v := range v1 {
		acc += math.Pow((v - v2[i]), 2)
	}
	return math.Pow(acc, 0.5)
}

// TODO: add other distance function options...
