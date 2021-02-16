package state

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
)

type PointShare struct {
	I kyber.Scalar
	P kyber.Point
}

func (ps *PointShare) Serialize() (buf []byte) {

	buf, err := ps.I.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}
	buf2, err := ps.P.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}
	buf = append(buf, buf2...)
	return
}

func (ps *PointShare) Deserialize(btes []byte) (*PointShare, error) {
	if len(btes) != 96 {
		return nil, fmt.Errorf("Wrong buffer length", len(btes))
	}
	ps.I = CurrentState.suite.G1().Scalar()
	ps.P = CurrentState.suite.G1().Point()
	ps.I.SetBytes(btes[:32])
	ps.P.UnmarshalBinary(btes[32:])
	return ps, nil
}

func (ps *PointShare) MarshalJSON() ([]byte, error) {
	bt := ps.Serialize()
	return []byte(hex.EncodeToString(bt)), nil
}

func (ps *PointShare) UnmarshalJSON(in []byte) error {
	ser, err := hex.DecodeString(string(in))
	if err != nil {
		return err
	}
	ps.Deserialize(ser)
	return nil
}

func PointShares(poly share.PriPoly, partsCount int) []*PointShare {
	shares := poly.Shares(partsCount)

	return PriShares2PointShares(shares)

}

func PriShares2PointShares(shares []*share.PriShare) []*PointShare {
	pshares := make([]*PointShare, 0, len(shares))

	for _, s := range shares {

		pshares = append(pshares, &PointShare{CurrentState.suite.G1().Scalar().SetInt64(int64(s.I + 1)), CurrentState.suite.G1().Point().Mul(s.V, nil)})

	}
	return pshares
}

func Mock() {

	secretScalar := CurrentState.suite.G1().Scalar().SetInt64(int64(42))
	fmt.Println("Secret number:", secretScalar)
	secretPoint := CurrentState.suite.G1().Point().Mul(secretScalar, nil)
	fmt.Println("Secret point:", secretPoint)

	T := 4

	poly := share.NewPriPoly(pairing.NewSuiteBn256(), T, secretScalar, pairing.NewSuiteBn256().RandomStream())
	fmt.Println(poly.Coefficients())
	shares := poly.Shares(6)
	wshares := WildShares(poly, RandomInts(6))
	for i, s := range shares {
		fmt.Println(i, *s)
	}

	for i, s := range wshares {
		fmt.Println(i, *s)
		b, e := s.MarshalJSON()
		fmt.Println(e, "JSON", string(b))
		s2 := new(PointShare)
		e = s2.UnmarshalJSON(b)
		fmt.Println(e, s2.P.Equal(s.P), s2.I.Equal(s.I))

	}

	rec, _ := share.RecoverSecret(CurrentState.suite.G1(), shares, T, T)
	b, _ := rec.MarshalBinary()
	r := new(big.Int)
	r.SetBytes(b)
	fmt.Println("Recovered scalar:", r)

	pshares := PriShares2PointShares(shares)
	rpt, err := RecoverSecretPoint(CurrentState.suite.G1(), pshares[:4], T)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Recovered point:", rpt)

	rpt, err = RecoverSecretPoint(CurrentState.suite.G1(), wshares[2:], T)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Recovered point:", rpt)

	badshares := append(pshares[:3], pshares[0])
	rpt, err = RecoverSecretPoint(CurrentState.suite.G1(), badshares, T)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Recovered (bad) point:", rpt)

}

//Now this is silly, but because of the poly.Eval() implementation the polynomila will be evaluated
//at evalpoints[i]+1...
func WildShares(poly *share.PriPoly, evalpoints []int) []*PointShare {
	shares := make([]*PointShare, 0, len(evalpoints))
	for _, ep := range evalpoints {
		ps := new(PointShare)
		ps.I = CurrentState.suite.G1().Scalar().SetInt64(int64(ep) + 1)
		ps.P = CurrentState.suite.G1().Point().Mul(poly.Eval(ep).V, nil)
		shares = append(shares, ps)
	}
	return shares
}

//Simply generating a alice of n pseudo-random numbers
//We do not even need to worry about 0, as the evaluation point Eval(v) is v+1
func RandomInts(n int) []int {
	rn := make([]int, n, n)
	for i := 0; i < n; i++ {
		rn[i] = rand.Int()
	}
	return rn
}

func RecoverSecretPoint(g kyber.Group, shares []*PointShare, t int) (kyber.Point, error) {

	if len(shares) < t {
		return nil, fmt.Errorf("share: not enough shares to recover secret")
	}

	acc := g.Point().Mul(g.Scalar().Zero(), nil)
	num := g.Point()
	den := g.Scalar()
	tmp := g.Scalar()

	for i, si := range shares {
		num.Set(si.P)
		den.One()

		for j, sj := range shares {
			if i == j {
				continue
			}

			num.Mul(sj.I, num)
			tmp.Sub(sj.I, si.I)
			den.Mul(den, tmp)
		}

		acc.Add(acc, g.Point().Mul(den.Inv(den), num))
	}

	return acc, nil
}
