package rand

import (
	"crypto/sha3"
	"math/rand"
	"mayo-threshold-go/model"
)

func SampleFieldElement() byte {
	randomInt := rand.Int()
	return byte(randomInt) & 0xf

	/*
		buf := make([]byte, 1) // TODO: Use crypto/rand when done

		_, err := rand.Read(buf)
		if err != nil {
			panic(err)
		}

		return buf[0] & 0xf
	*/
}

func Shake256(outputLength int, inputs ...[]byte) []byte {
	output := make([]byte, outputLength)

	h := sha3.NewSHAKE256()
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(output[:])

	for index, elem := range output {
		output[index] = elem & 0xf
	}

	return output
}

func Coin(parties []*model.Party, lambda int) []byte {
	result := make([]byte, lambda+64)

	for i := 0; i < lambda+64; i++ {
		for _, _ = range parties {
			result[i] ^= SampleFieldElement()
		}
	}

	return result
}

func CoinMatrix(parties []*model.Party, r, c int) [][]byte {
	matrix := make([][]byte, r)
	for i := range matrix {
		matrix[i] = make([]byte, c)
		for j := 0; j < c; j++ {
			for range parties {
				matrix[i][j] ^= SampleFieldElement()
			}
		}
	}
	return matrix
}

func Matrix(r, c int) [][]byte {
	result := make([][]byte, r)

	for i := 0; i < r; i++ {
		row := make([]byte, c)
		for j := 0; j < c; j++ {
			row[j] = SampleFieldElement()
		}
		result[i] = row
	}

	return result
}

func Vector(c int) []byte {
	result := make([]byte, c)

	for i := 0; i < c; i++ {
		result[i] = SampleFieldElement()
	}

	return result
}
