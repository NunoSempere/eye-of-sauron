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

import "math"

func max(values []float64) float64 {
	maxNum := float64(math.MinInt64)
	for _, v := range values {
		if v > maxNum {
			maxNum = v
		}
	}

	return maxNum
}

func min(values []float64) (int, float64) {
	minNum := math.MaxFloat64
	var index int
	for i, v := range values {
		if v < minNum {
			index = i
			minNum = v
		}
	}

	return index, minNum
}

func unfold(data [][]float64) []float64 {
	var ud []float64
	for _, row := range data {
		ud = append(ud, row...)
	}
	return ud
}

func isNaN(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}

	return value
}

func isInf(value float64) float64 {
	if math.IsInf(value, 1) {
		return math.MaxFloat64
	}

	if math.IsInf(value, -1) {
		return float64(math.MinInt64)
	}

	return value
}

func isNum(value float64) float64 {
	return isNaN(isInf(value))
}

func containsNode(list []node, n node) bool {
	for _, node := range list {
		if node.key == n.key {
			return true
		}
	}

	return false
}

func containsInt(list []int, i int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}

	return false
}
