// Copyright 2020 ConsenSys Software Inc.
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

// Code generated by gnark DO NOT EDIT

package cs

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// hintFunction signature hints functions must match
type hintFunction func(input []fr.Element) (fr.Element, error)

// qMinusOne is used by powModulusMinusOne
var qMinusOne big.Int

func init() {
	qMinusOne.SetUint64(1)
	qMinusOne.Sub(fr.Modulus(), &qMinusOne)
}

// powModulusMinusOne expects len(inputs) == 1
// inputs[0] == a
// returns m = a^(modulus-1) - 1
func powModulusMinusOne(inputs []fr.Element) (fr.Element, error) {
	if len(inputs) != 1 {
		return fr.Element{}, errors.New("powModulusMinusOne expects one input")
	}
	var v fr.Element
	v.Exp(inputs[0], &qMinusOne)
	one := fr.One()
	v.Sub(&one, &v)
	return v, nil
}

// ithBit expects len(inputs) == 2
// inputs[0] == a
// inputs[1] == n
// returns bit number n of a
func ithBit(inputs []fr.Element) (fr.Element, error) {
	if len(inputs) != 2 {
		return fr.Element{}, errors.New("ithBit expects 2 inputs; inputs[0] == value, inputs[1] == bit position")
	}
	// TODO @gbotrel this is very inneficient; it adds ~256*2 multiplications to extract all bits of a value.
	inputs[0].FromMont()
	inputs[1].FromMont()
	if !inputs[1].IsUint64() {
		return fr.Element{}, errors.New("ithBit expects bit position (input[1]) to fit on one word")
	}
	if inputs[0].Bit(inputs[1][0]) == 0 {
		return fr.Element{}, nil
	}
	return fr.One(), nil
}
