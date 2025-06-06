package mpc

import (
	"fmt"
	"mayo-threshold-go/model"
	"mayo-threshold-go/rand"
)

func (c *Context) computeT(parties []*model.Party, iteration int) bool {
	s := len(parties[0].A)
	t := len(parties[0].A[0])

	SShares := c.algo.createSharesForRandomMatrix(t, t)
	RShares := c.algo.createSharesForRandomMatrix(s, s)
	for partyNumber, party := range parties {
		party.S = SShares[partyNumber]
		party.R = RShares[partyNumber]
	}

	// Compute [A * S] = [A] * [S]
	dShares := make([][][]byte, len(parties))
	eShares := make([][][]byte, len(parties))
	for partyNumber, party := range parties {
		ai := c.signTriples.ComputeT1[iteration].A[partyNumber]
		bi := c.signTriples.ComputeT1[iteration].B[partyNumber]
		di := AddMatricesNew(party.A, ai)
		ei := AddMatricesNew(party.S, bi)

		dShares[partyNumber] = di
		eShares[partyNumber] = ei
	}
	ATimesSShares := c.multiplicationProtocol(parties, c.signTriples.ComputeT1[iteration], dShares, eShares)

	// Compute [T] = [R] * [A * S]
	dShares = make([][][]byte, len(parties))
	eShares = make([][][]byte, len(parties))
	for partyNumber, party := range parties {
		ai := c.signTriples.ComputeT2[iteration].A[partyNumber]
		bi := c.signTriples.ComputeT2[iteration].B[partyNumber]
		di := AddMatricesNew(party.R, ai)
		ei := AddMatricesNew(ATimesSShares[partyNumber], bi)

		dShares[partyNumber] = di
		eShares[partyNumber] = ei
	}
	TShares := c.multiplicationProtocol(parties, c.signTriples.ComputeT2[iteration], dShares, eShares)

	// Open T and check rank
	T := c.algo.openMatrix(TShares)

	for _, party := range parties {
		party.T = T
	}

	return isFullRank(T)
}

func (c *Context) computeAInverse(parties []*model.Party) {
	dShares := make([][][]byte, len(parties))
	eShares := make([][][]byte, len(parties))

	// Compute locally
	for partyNumber, party := range parties {
		ai := c.signTriples.ComputeAInverse.A[partyNumber]
		bi := c.signTriples.ComputeAInverse.B[partyNumber]
		TInverse := computeRightInverse(party.T)

		di := AddMatricesNew(MultiplyMatrices(party.S, TInverse), ai)
		ei := AddMatricesNew(party.R, bi)

		dShares[partyNumber] = di
		eShares[partyNumber] = ei
	}

	// Open d, e and compute locally
	zShares := c.multiplicationProtocol(parties, c.signTriples.ComputeAInverse, dShares, eShares)
	for partyNumber, party := range parties {
		party.AInverse = zShares[partyNumber]
	}
}

func (c *Context) computeLittleX(parties []*model.Party) {
	s := len(parties[0].A)
	t := len(parties[0].A[0])

	basis := computeBasisOfKernel(parties[0].T)

	if len(basis) != t-s {
		panic(fmt.Errorf("length of basis is incorrect"))
	}

	for _, party := range parties {
		z := make([]byte, t)
		zVector := rand.Vector(t - s)

		for index, elem := range zVector {
			z = AddVec(z, MultiplyVecConstant(elem, basis[index]))
		}

		party.Z = z
	}

	// Compute [A^-1] * [b]
	dShares := make([][][]byte, len(parties))
	eShares := make([][][]byte, len(parties))
	for partyNumber, party := range parties {
		ai := c.signTriples.ComputeX1.A[partyNumber]
		bi := c.signTriples.ComputeX1.B[partyNumber]
		di := AddMatricesNew(party.AInverse, ai)
		ei := AddMatricesNew(vectorToMatrix(party.LittleY), bi)

		dShares[partyNumber] = di
		eShares[partyNumber] = ei
	}
	AInvTimesB := c.multiplicationProtocol(parties, c.signTriples.ComputeX1, dShares, eShares)

	// Compute [S] * [z]
	dShares = make([][][]byte, len(parties))
	eShares = make([][][]byte, len(parties))
	for partyNumber, party := range parties {
		ai := c.signTriples.ComputeX2.A[partyNumber]
		bi := c.signTriples.ComputeX2.B[partyNumber]
		di := AddMatricesNew(party.S, ai)
		ei := AddMatricesNew(vectorToMatrix(party.Z), bi)

		dShares[partyNumber] = di
		eShares[partyNumber] = ei
	}
	STimesZ := c.multiplicationProtocol(parties, c.signTriples.ComputeX2, dShares, eShares)

	// [x] = [A^-1] * [b] + [S] * [z]
	for i, party := range parties {
		party.LittleX = matrixToVec(AddMatricesNew(AInvTimesB[i], STimesZ[i]))
	}
}
