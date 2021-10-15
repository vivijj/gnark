/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sw

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/fields"
	"github.com/consensys/gnark/test"
)

type pairingBLS377 struct {
	P          G1Affine `gnark:",public"`
	Q          G2Affine
	pairingRes bls12377.GT
}

func (circuit *pairingBLS377) Define(curveID ecc.ID, gnark frontend.API) error {

	ateLoop := uint64(9586122913090633729)
	ext := fields.GetBLS377ExtensionFp12(gnark)
	pairingInfo := PairingContext{AteLoop: ateLoop, Extension: ext}
	pairingInfo.BTwistCoeff.A0 = gnark.Constant(0)
	pairingInfo.BTwistCoeff.A1 = gnark.Constant("155198655607781456406391640216936120121836107652948796323930557600032281009004493664981332883744016074664192874906")

	milRes := fields.E12{}
	MillerLoop(gnark, circuit.P, circuit.Q, &milRes, pairingInfo)

	pairingRes := fields.E12{}
	pairingRes.FinalExponentiation(gnark, &milRes, ateLoop, ext)

	mustbeEq(gnark, pairingRes, &circuit.pairingRes)

	return nil
}

func TestPairingBLS377(t *testing.T) {

	// pairing test data
	P, Q, pairingRes := pairingData()

	// create cs
	var circuit, witness pairingBLS377
	circuit.pairingRes = pairingRes

	// assign values to witness
	witness.P.Assign(&P)
	witness.Q.Assign(&Q)

	assert := test.NewAssert(t)
	assert.SolvingSucceeded(&circuit, &witness, test.WithCurves(ecc.BW6_761), test.WithBackends(backend.GROTH16))

}

func pairingData() (P bls12377.G1Affine, Q bls12377.G2Affine, pairingRes bls12377.GT) {
	_, _, P, Q = bls12377.Generators()
	milRes, _ := bls12377.MillerLoop([]bls12377.G1Affine{P}, []bls12377.G2Affine{Q})
	pairingRes = bls12377.FinalExponentiation(&milRes)
	return
}

func mustbeEq(gnark frontend.API, fp12 fields.E12, e12 *bls12377.GT) {
	gnark.AssertIsEqual(fp12.C0.B0.A0, e12.C0.B0.A0)
	gnark.AssertIsEqual(fp12.C0.B0.A1, e12.C0.B0.A1)
	gnark.AssertIsEqual(fp12.C0.B1.A0, e12.C0.B1.A0)
	gnark.AssertIsEqual(fp12.C0.B1.A1, e12.C0.B1.A1)
	gnark.AssertIsEqual(fp12.C0.B2.A0, e12.C0.B2.A0)
	gnark.AssertIsEqual(fp12.C0.B2.A1, e12.C0.B2.A1)
	gnark.AssertIsEqual(fp12.C1.B0.A0, e12.C1.B0.A0)
	gnark.AssertIsEqual(fp12.C1.B0.A1, e12.C1.B0.A1)
	gnark.AssertIsEqual(fp12.C1.B1.A0, e12.C1.B1.A0)
	gnark.AssertIsEqual(fp12.C1.B1.A1, e12.C1.B1.A1)
	gnark.AssertIsEqual(fp12.C1.B2.A0, e12.C1.B2.A0)
	gnark.AssertIsEqual(fp12.C1.B2.A1, e12.C1.B2.A1)
}
