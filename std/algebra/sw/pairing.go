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
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/fields"
)

// PairingContext contains useful info about the pairing
type PairingContext struct {
	AteLoop     uint64 // stores the ate loop
	Extension   fields.Extension
	BTwistCoeff fields.E2
}

// lineEvaluation represents a sparse Fp12 Elmt (result of the line evaluation)
type lineEvaluation struct {
	r0, r1, r2 fields.E2
}

// MillerLoop computes the miller loop
func MillerLoop(gnark frontend.API, P G1Affine, Q G2Affine, res *fields.E12, pairingInfo PairingContext) *fields.E12 {

	var ateLoopBin [64]uint
	var ateLoopBigInt big.Int
	ateLoopBigInt.SetUint64(pairingInfo.AteLoop)
	for i := 0; i < 64; i++ {
		ateLoopBin[i] = ateLoopBigInt.Bit(i)
	}

	res.SetOne(gnark)
	var l lineEvaluation

	var qProj G2Proj
	qProj.X = Q.X
	qProj.Y = Q.Y
	qProj.Z.A0 = gnark.Constant(1)
	qProj.Z.A1 = gnark.Constant(0)

	// Miller loop
	for i := len(ateLoopBin) - 2; i >= 0; i-- {

		// res <- res**2
		res.Mul(gnark, res, res, pairingInfo.Extension)

		// l(P) where div(l) = 2(qProj)+([-2]qProj)-2(O)
		// qProj <- 2*qProj
		qProj.DoubleStep(gnark, &l, pairingInfo)
		l.r0.MulByFp(gnark, &l.r0, P.Y)
		l.r1.MulByFp(gnark, &l.r1, P.X)

		// res <- res*l(P)
		res.MulBy034(gnark, &l.r0, &l.r1, &l.r2, pairingInfo.Extension)

		if ateLoopBin[i] == 0 {
			continue
		}

		// l(P) where div(l) = (qProj)+(Q)+(-Q-qProj)-3(O)
		// qProj <- qProj + Q
		qProj.AddMixedStep(gnark, &l, &Q, pairingInfo)
		l.r0.MulByFp(gnark, &l.r0, P.Y)
		l.r1.MulByFp(gnark, &l.r1, P.X)

		// res <- res*l(P)
		res.MulBy034(gnark, &l.r0, &l.r1, &l.r2, pairingInfo.Extension)

	}

	return res
}

// DoubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *G2Proj) DoubleStep(gnark frontend.API, evaluation *lineEvaluation, pairingInfo PairingContext) {

	// get some Element from our pool
	var t0, t1, A, B, C, D, E, EE, F, G, H, I, J, K fields.E2
	twoInv := gnark.Constant(2)
	twoInv = gnark.Inverse(twoInv)
	t0.Mul(gnark, &p.X, &p.Y, pairingInfo.Extension)
	A.MulByFp(gnark, &t0, twoInv)
	B.Mul(gnark, &p.Y, &p.Y, pairingInfo.Extension)
	C.Mul(gnark, &p.Z, &p.Z, pairingInfo.Extension)
	D.Add(gnark, &C, &C).
		Add(gnark, &D, &C)
	E.Mul(gnark, &D, &pairingInfo.BTwistCoeff, pairingInfo.Extension)
	F.Add(gnark, &E, &E).
		Add(gnark, &F, &E)
	G.Add(gnark, &B, &F)
	G.MulByFp(gnark, &G, twoInv)
	H.Add(gnark, &p.Y, &p.Z).
		Mul(gnark, &H, &H, pairingInfo.Extension)
	t1.Add(gnark, &B, &C)
	H.Sub(gnark, &H, &t1)
	I.Sub(gnark, &E, &B)
	J.Mul(gnark, &p.X, &p.X, pairingInfo.Extension)
	EE.Mul(gnark, &E, &E, pairingInfo.Extension)
	K.Add(gnark, &EE, &EE).
		Add(gnark, &K, &EE)

	// X, Y, Z
	p.X.Sub(gnark, &B, &F).
		Mul(gnark, &p.X, &A, pairingInfo.Extension)
	p.Y.Mul(gnark, &G, &G, pairingInfo.Extension).
		Sub(gnark, &p.Y, &K)
	p.Z.Mul(gnark, &B, &H, pairingInfo.Extension)

	// Line evaluation
	evaluation.r0.Neg(gnark, &H)
	evaluation.r1.Add(gnark, &J, &J).
		Add(gnark, &evaluation.r1, &J)
	evaluation.r2 = I
}

// AddMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *G2Proj) AddMixedStep(gnark frontend.API, evaluation *lineEvaluation, a *G2Affine, pairingInfo PairingContext) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, C, D, E, F, G, H, t0, t1, t2, J fields.E2
	Y2Z1.Mul(gnark, &a.Y, &p.Z, pairingInfo.Extension)
	O.Sub(gnark, &p.Y, &Y2Z1)
	X2Z1.Mul(gnark, &a.X, &p.Z, pairingInfo.Extension)
	L.Sub(gnark, &p.X, &X2Z1)
	C.Mul(gnark, &O, &O, pairingInfo.Extension)
	D.Mul(gnark, &L, &L, pairingInfo.Extension)
	E.Mul(gnark, &L, &D, pairingInfo.Extension)
	F.Mul(gnark, &p.Z, &C, pairingInfo.Extension)
	G.Mul(gnark, &p.X, &D, pairingInfo.Extension)
	t0.Add(gnark, &G, &G)
	H.Add(gnark, &E, &F).
		Sub(gnark, &H, &t0)
	t1.Mul(gnark, &p.Y, &E, pairingInfo.Extension)

	// X, Y, Z
	p.X.Mul(gnark, &L, &H, pairingInfo.Extension)
	p.Y.Sub(gnark, &G, &H).
		Mul(gnark, &p.Y, &O, pairingInfo.Extension).
		Sub(gnark, &p.Y, &t1)
	p.Z.Mul(gnark, &E, &p.Z, pairingInfo.Extension)

	t2.Mul(gnark, &L, &a.Y, pairingInfo.Extension)
	J.Mul(gnark, &a.X, &O, pairingInfo.Extension).
		Sub(gnark, &J, &t2)

	// Line evaluation
	evaluation.r0 = L
	evaluation.r1.Neg(gnark, &O)
	evaluation.r2 = J
}
