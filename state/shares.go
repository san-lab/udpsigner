package state

import (
	"fmt"
	"math/big"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
)

type PointShare struct {
	I int
	P kyber.Point
}

func PointShares(poly share.PriPoly, partsCount int) []*PointShare {
	shares := poly.Shares(partsCount)

	return PriShares2PointShares(shares)

}

func PriShares2PointShares(shares []*share.PriShare) []*PointShare {
	pshares := make([]*PointShare, 0, len(shares))

	for _, s := range shares {
		pshares = append(pshares, &PointShare{s.I, CurrentState.suite.G1().Point().Mul(s.V, nil)})

	}
	for _, ps := range pshares {
		fmt.Println(*ps)
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
	shares := poly.Shares(4)
	for i, s := range shares {
		fmt.Println(i, *s)
	}

	rec, _ := share.RecoverSecret(CurrentState.suite.G1(), shares, T, T)
	b, _ := rec.MarshalBinary()
	r := new(big.Int)
	r.SetBytes(b)
	fmt.Println("Recovered scalar:", r)

	rpt0, _ := RecoverSecretP2(CurrentState.suite.G1(), shares, T, T)
	fmt.Println("Recovered point:", rpt0)

	pshares := PriShares2PointShares(shares)
	rpt, _ := RecoverSecretPoint(CurrentState.suite.G1(), pshares, T)
	fmt.Println("Recovered point:", rpt)

}

func RecoverSecretPoint(g kyber.Group, shares []*PointShare, t int) (kyber.Point, error) {

	if len(shares) < t {
		return nil, fmt.Errorf("share: not enough shares to recover secret")
	}

	acc := g.Point().Mul(g.Scalar().Zero(), nil)
	num := g.Point()
	den := g.Scalar()

	for i, si := range shares {
		num.Set(si.P)
		den.One()
		xi := g.Scalar().SetInt64(int64(si.I + 1))
		for j, sj := range shares {
			if i == j {
				continue
			}
			xj := g.Scalar().SetInt64(int64(sj.I + 1))
			num.Mul(xj, num)
			den.Sub(xj, xi)
			den.Inv(den)
		}
		acc.Add(acc, g.Point().Mul(den, num))
	}

	return acc, nil
}

func RecoverSecretP2(g kyber.Group, shares []*share.PriShare, t, n int) (kyber.Point, error) {
	x, y := xyScalar(g, shares, t, n)
	if len(x) < t {
		return nil, fmt.Errorf("share: not enough shares to recover secret")
	}

	acc := g.Point().Mul(g.Scalar().Zero(), nil)
	num := g.Point()
	den := g.Scalar()
	tmp := g.Scalar()

	for i, xi := range x {
		yi := y[i]
		num.Mul(yi, nil)
		den.One()
		for j, xj := range x {
			if i == j {
				continue
			}
			num.Mul(xj, num)
			den.Mul(den, tmp.Sub(xj, xi))
		}
		acc.Add(acc, num.Mul(den.Inv(den), num))
	}

	return acc, nil
}

type byIndexScalar []*share.PriShare

func (s byIndexScalar) Len() int           { return len(s) }
func (s byIndexScalar) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byIndexScalar) Less(i, j int) bool { return s[i].I < s[j].I }

// xyScalar returns the list of (x_i, y_i) pairs indexed. The first map returned
// is the list of x_i and the second map is the list of y_i, both indexed in
// their respective map at index i.
func xyScalar(g kyber.Group, shares []*share.PriShare, t, n int) (map[int]kyber.Scalar, map[int]kyber.Scalar) {
	// we are sorting first the shares since the shares may be unrelated for
	// some applications. In this case, all participants needs to interpolate on
	// the exact same order shares.
	sorted := make([]*share.PriShare, 0, n)
	for _, share := range shares {
		if share != nil {
			sorted = append(sorted, share)
		}
	}
	//sort.Sort(byIndexScalar(sorted))

	x := make(map[int]kyber.Scalar)
	y := make(map[int]kyber.Scalar)
	for _, s := range sorted {
		if s == nil || s.V == nil || s.I < 0 {
			continue
		}
		idx := s.I
		x[idx] = g.Scalar().SetInt64(int64(idx + 1))
		y[idx] = s.V
		if len(x) == t {
			break
		}
	}
	return x, y
}
