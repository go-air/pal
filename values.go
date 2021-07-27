// Copyright 2021 Scott Cotton
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pal


type ValueKind int

const (
	Const ValueKind = iota
	Var

)

type Value 


type Values[V any]  interface {
	Const(v V) (uint64, bool)
	Var(v V) bool
	Plus(v V) Value
	Equal(a, b V) AbsTruth
	Less(a, b V) AbsTruth
}
