// Copyright 2020 ConsenSys AG
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

// Package groth16 implements Groth16 zkSNARK workflow (https://eprint.iacr.org/2016/260.pdf)
package groth16

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc"

	"github.com/consensys/gnark/frontend"
	backend_bls377 "github.com/consensys/gnark/internal/backend/bls12-377/cs"
	backend_bls381 "github.com/consensys/gnark/internal/backend/bls12-381/cs"
	backend_bn256 "github.com/consensys/gnark/internal/backend/bn254/cs"
	backend_bw761 "github.com/consensys/gnark/internal/backend/bw6-761/cs"

	witness_bls377 "github.com/consensys/gnark/internal/backend/bls12-377/witness"
	witness_bls381 "github.com/consensys/gnark/internal/backend/bls12-381/witness"
	witness_bn256 "github.com/consensys/gnark/internal/backend/bn254/witness"
	witness_bw761 "github.com/consensys/gnark/internal/backend/bw6-761/witness"

	gnarkio "github.com/consensys/gnark/io"

	groth16_bls377 "github.com/consensys/gnark/internal/backend/bls12-377/groth16"
	groth16_bls381 "github.com/consensys/gnark/internal/backend/bls12-381/groth16"
	groth16_bn256 "github.com/consensys/gnark/internal/backend/bn254/groth16"
	groth16_bw761 "github.com/consensys/gnark/internal/backend/bw6-761/groth16"
)

// Proof represents a Groth16 proof generated by groth16.Prove
//
// it's underlying implementation is curve specific (see gnark/internal/backend)
type Proof interface {
	gnarkio.WriterRawTo
	io.WriterTo
	io.ReaderFrom
}

// ProvingKey represents a Groth16 ProvingKey
//
// it's underlying implementation is curve specific (see gnark/internal/backend)
type ProvingKey interface {
	gnarkio.WriterRawTo
	io.WriterTo
	io.ReaderFrom
	IsDifferent(interface{}) bool
}

// VerifyingKey represents a Groth16 VerifyingKey
//
// it's underlying implementation is curve specific (see gnark/internal/backend)
//
// ExportSolidity is implemented for BN254 and will return an error with other curves
type VerifyingKey interface {
	gnarkio.WriterRawTo
	io.WriterTo
	io.ReaderFrom
	SizePublicWitness() int // number of elements expected in the public witness
	IsDifferent(interface{}) bool
	ExportSolidity(w io.Writer) error
}

// Verify runs the groth16.Verify algorithm on provided proof with given witness
func Verify(proof Proof, vk VerifyingKey, publicWitness frontend.Circuit) error {

	switch _proof := proof.(type) {
	case *groth16_bls377.Proof:
		w := witness_bls377.Witness{}
		if err := w.FromPublicAssignment(publicWitness); err != nil {
			return err
		}
		return groth16_bls377.Verify(_proof, vk.(*groth16_bls377.VerifyingKey), w)
	case *groth16_bls381.Proof:
		w := witness_bls381.Witness{}
		if err := w.FromPublicAssignment(publicWitness); err != nil {
			return err
		}
		return groth16_bls381.Verify(_proof, vk.(*groth16_bls381.VerifyingKey), w)
	case *groth16_bn256.Proof:
		w := witness_bn256.Witness{}
		if err := w.FromPublicAssignment(publicWitness); err != nil {
			return err
		}
		return groth16_bn256.Verify(_proof, vk.(*groth16_bn256.VerifyingKey), w)
	case *groth16_bw761.Proof:
		w := witness_bw761.Witness{}
		if err := w.FromPublicAssignment(publicWitness); err != nil {
			return err
		}
		return groth16_bw761.Verify(_proof, vk.(*groth16_bw761.VerifyingKey), w)
	default:
		panic("unrecognized R1CS curve type")
	}
}

// ReadAndVerify behaves like Verify, except witness is read from a io.Reader
// witness must be [uint32(nbElements) | publicVariables ]
func ReadAndVerify(proof Proof, vk VerifyingKey, publicWitness io.Reader) error {

	switch _vk := vk.(type) {
	case *groth16_bls377.VerifyingKey:
		w := witness_bls377.Witness{}
		if _, err := w.LimitReadFrom(publicWitness, vk.SizePublicWitness()); err != nil {
			return err
		}
		return groth16_bls377.Verify(proof.(*groth16_bls377.Proof), _vk, w)
	case *groth16_bls381.VerifyingKey:
		w := witness_bls381.Witness{}
		if _, err := w.LimitReadFrom(publicWitness, vk.SizePublicWitness()); err != nil {
			return err
		}
		return groth16_bls381.Verify(proof.(*groth16_bls381.Proof), _vk, w)
	case *groth16_bn256.VerifyingKey:
		w := witness_bn256.Witness{}
		if _, err := w.LimitReadFrom(publicWitness, vk.SizePublicWitness()); err != nil {
			return err
		}
		return groth16_bn256.Verify(proof.(*groth16_bn256.Proof), _vk, w)
	case *groth16_bw761.VerifyingKey:
		w := witness_bw761.Witness{}
		if _, err := w.LimitReadFrom(publicWitness, vk.SizePublicWitness()); err != nil {
			return err
		}
		return groth16_bw761.Verify(proof.(*groth16_bw761.Proof), _vk, w)
	default:
		panic("unrecognized R1CS curve type")
	}
}

// Prove generates the proof of knoweldge of a r1cs with witness.
// if force flag is set, Prove ignores R1CS solving error (ie invalid witness) and executes
// the FFTs and MultiExponentiations to compute an (invalid) Proof object
func Prove(r1cs frontend.CompiledConstraintSystem, pk ProvingKey, witness frontend.Circuit, force ...bool) (Proof, error) {

	_force := false
	if len(force) > 0 {
		_force = force[0]
	}

	switch _r1cs := r1cs.(type) {
	case *backend_bls377.R1CS:
		w := witness_bls377.Witness{}
		if err := w.FromFullAssignment(witness); err != nil {
			return nil, err
		}
		return groth16_bls377.Prove(_r1cs, pk.(*groth16_bls377.ProvingKey), w, _force)
	case *backend_bls381.R1CS:
		w := witness_bls381.Witness{}
		if err := w.FromFullAssignment(witness); err != nil {
			return nil, err
		}
		return groth16_bls381.Prove(_r1cs, pk.(*groth16_bls381.ProvingKey), w, _force)
	case *backend_bn256.R1CS:
		w := witness_bn256.Witness{}
		if err := w.FromFullAssignment(witness); err != nil {
			return nil, err
		}
		return groth16_bn256.Prove(_r1cs, pk.(*groth16_bn256.ProvingKey), w, _force)
	case *backend_bw761.R1CS:
		w := witness_bw761.Witness{}
		if err := w.FromFullAssignment(witness); err != nil {
			return nil, err
		}
		return groth16_bw761.Prove(_r1cs, pk.(*groth16_bw761.ProvingKey), w, _force)
	default:
		panic("unrecognized R1CS curve type")
	}
}

// ReadAndProve behaves like Prove, except witness is read from a io.Reader
// witness must be [uint32(nbElements) | publicVariables | secretVariables]
func ReadAndProve(r1cs frontend.CompiledConstraintSystem, pk ProvingKey, witness io.Reader, force ...bool) (Proof, error) {
	_force := false
	if len(force) > 0 {
		_force = force[0]
	}

	_, nbSecret, nbPublic := r1cs.GetNbVariables()
	expectedSize := (nbSecret + nbPublic - 1)

	switch _r1cs := r1cs.(type) {
	case *backend_bls377.R1CS:
		w := witness_bls377.Witness{}
		if _, err := w.LimitReadFrom(witness, expectedSize); err != nil {
			return nil, err
		}
		return groth16_bls377.Prove(_r1cs, pk.(*groth16_bls377.ProvingKey), w, _force)
	case *backend_bls381.R1CS:
		w := witness_bls381.Witness{}
		if _, err := w.LimitReadFrom(witness, expectedSize); err != nil {
			return nil, err
		}
		return groth16_bls381.Prove(_r1cs, pk.(*groth16_bls381.ProvingKey), w, _force)
	case *backend_bn256.R1CS:
		w := witness_bn256.Witness{}
		if _, err := w.LimitReadFrom(witness, expectedSize); err != nil {
			return nil, err
		}
		return groth16_bn256.Prove(_r1cs, pk.(*groth16_bn256.ProvingKey), w, _force)
	case *backend_bw761.R1CS:
		w := witness_bw761.Witness{}
		if _, err := w.LimitReadFrom(witness, expectedSize); err != nil {
			return nil, err
		}
		return groth16_bw761.Prove(_r1cs, pk.(*groth16_bw761.ProvingKey), w, _force)
	default:
		panic("unrecognized R1CS curve type")
	}
}

// Setup runs groth16.Setup with provided R1CS
func Setup(r1cs frontend.CompiledConstraintSystem) (ProvingKey, VerifyingKey, error) {

	switch _r1cs := r1cs.(type) {
	case *backend_bls377.R1CS:
		var pk groth16_bls377.ProvingKey
		var vk groth16_bls377.VerifyingKey
		if err := groth16_bls377.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *backend_bls381.R1CS:
		var pk groth16_bls381.ProvingKey
		var vk groth16_bls381.VerifyingKey
		if err := groth16_bls381.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *backend_bn256.R1CS:
		var pk groth16_bn256.ProvingKey
		var vk groth16_bn256.VerifyingKey
		if err := groth16_bn256.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *backend_bw761.R1CS:
		var pk groth16_bw761.ProvingKey
		var vk groth16_bw761.VerifyingKey
		if err := groth16_bw761.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	default:
		panic("unrecognized R1CS curve type")
	}
}

// DummySetup create a random ProvingKey with provided R1CS
// it doesn't return a VerifyingKey and is use for benchmarking or test purposes only.
func DummySetup(r1cs frontend.CompiledConstraintSystem) (ProvingKey, error) {
	switch _r1cs := r1cs.(type) {
	case *backend_bls377.R1CS:
		var pk groth16_bls377.ProvingKey
		if err := groth16_bls377.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *backend_bls381.R1CS:
		var pk groth16_bls381.ProvingKey
		if err := groth16_bls381.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *backend_bn256.R1CS:
		var pk groth16_bn256.ProvingKey
		if err := groth16_bn256.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *backend_bw761.R1CS:
		var pk groth16_bw761.ProvingKey
		if err := groth16_bw761.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	default:
		panic("unrecognized R1CS curve type")
	}
}

// NewProvingKey instantiates a curve-typed ProvingKey and returns an interface object
// This function exists for serialization purposes
func NewProvingKey(curveID ecc.ID) ProvingKey {
	var pk ProvingKey
	switch curveID {
	case ecc.BN254:
		pk = &groth16_bn256.ProvingKey{}
	case ecc.BLS12_377:
		pk = &groth16_bls377.ProvingKey{}
	case ecc.BLS12_381:
		pk = &groth16_bls381.ProvingKey{}
	case ecc.BW6_761:
		pk = &groth16_bw761.ProvingKey{}
	default:
		panic("not implemented")
	}
	return pk
}

// NewVerifyingKey instantiates a curve-typed VerifyingKey and returns an interface
// This function exists for serialization purposes
func NewVerifyingKey(curveID ecc.ID) VerifyingKey {
	var vk VerifyingKey
	switch curveID {
	case ecc.BN254:
		vk = &groth16_bn256.VerifyingKey{}
	case ecc.BLS12_377:
		vk = &groth16_bls377.VerifyingKey{}
	case ecc.BLS12_381:
		vk = &groth16_bls381.VerifyingKey{}
	case ecc.BW6_761:
		vk = &groth16_bw761.VerifyingKey{}
	default:
		panic("not implemented")
	}

	return vk
}

// NewProof instantiates a curve-typed Proof and returns an interface
// This function exists for serialization purposes
func NewProof(curveID ecc.ID) Proof {
	var proof Proof
	switch curveID {
	case ecc.BN254:
		proof = &groth16_bn256.Proof{}
	case ecc.BLS12_377:
		proof = &groth16_bls377.Proof{}
	case ecc.BLS12_381:
		proof = &groth16_bls381.Proof{}
	case ecc.BW6_761:
		proof = &groth16_bw761.Proof{}
	default:
		panic("not implemented")
	}

	return proof
}

// NewCS instantiate a concrete curved-typed R1CS and return a R1CS interface
// This method exists for (de)serialization purposes
func NewCS(curveID ecc.ID) frontend.CompiledConstraintSystem {
	var r1cs frontend.CompiledConstraintSystem
	switch curveID {
	case ecc.BN254:
		r1cs = &backend_bn256.R1CS{}
	case ecc.BLS12_377:
		r1cs = &backend_bls377.R1CS{}
	case ecc.BLS12_381:
		r1cs = &backend_bls381.R1CS{}
	case ecc.BW6_761:
		r1cs = &backend_bw761.R1CS{}
	default:
		panic("not implemented")
	}
	return r1cs
}
