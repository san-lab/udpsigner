package state

import (
	"encoding/hex"
	"fmt"
	"math/rand"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
)

type ScalarShare struct {
	E       kyber.Scalar //Evaluation point
	V       kyber.Scalar //Value at E
	T       int          //Theshold of the Suite
	SuiteID string       //Suite ID
}

func (ss *ScalarShare) Serialize() (ser []byte) {

	//First T
	ser = []byte{byte(ss.T)}

	//Then Eval point
	buf1, err := ss.E.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}

	ser = append(ser, buf1...)

	//Then Point-Value
	buf2, err := ss.V.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}

	ser = append(ser, buf2...)

	//Last SuiteID
	ser = append(ser, []byte(ss.SuiteID)...)
	return
}

func (ss *ScalarShare) Deserialize(btes []byte) (*ScalarShare, error) {
	if len(btes) < 65 {
		return nil, fmt.Errorf("Wrong buffer length", len(btes))
	}
	ss.T = int(btes[0])
	ss.E = CurrentState.suite.G2().Scalar()
	ss.V = CurrentState.suite.G2().Scalar()

	ss.E.SetBytes(btes[1:33])
	ss.V.SetBytes(btes[33:65])
	if len(btes) > 97 {
		ss.SuiteID = string(btes[65:])
	}
	return ss, nil
}

//PointShare , as contrasted with kyber.share.PriShare, the value of E is NOT shifted by 1
//It also keeps the threshold value and the ID common to all matching shares
type PointShare struct {
	E       kyber.Scalar
	P       kyber.Point
	T       int
	SuiteID string
}

func ScalarShareToPointShare(scs *ScalarShare) *PointShare {
	ps := new(PointShare)
	ps.E = scs.E
	ps.SuiteID = scs.SuiteID
	ps.T = scs.T
	ps.P = CurrentState.suite.G2().Point().Mul(scs.V, nil)
	return ps
}

func (ps *PointShare) Serialize() (ser []byte) {

	//First T
	ser = []byte{byte(ps.T)}

	//Then Eval point
	buf1, err := ps.E.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}

	ser = append(ser, buf1...)

	//Then Point-Value
	buf2, err := ps.P.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		return
	}

	ser = append(ser, buf2...)

	//Last SuiteID
	ser = append(ser, []byte(ps.SuiteID)...)
	return
}

func (ps *PointShare) Deserialize(btes []byte) (*PointShare, error) {
	if len(btes) < 161 {
		return nil, fmt.Errorf("Wrong buffer length", len(btes))
	}
	ps.T = int(btes[0])
	ps.E = CurrentState.suite.G2().Scalar()
	ps.P = CurrentState.suite.G2().Point()

	ps.E.SetBytes(btes[1:33])
	ps.P.UnmarshalBinary(btes[33:161])
	if len(btes) > 97 {
		ps.SuiteID = string(btes[161:])
	}
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

//Now this is silly, but because of the poly.Eval() implementation the polynomila will be evaluated
//at evalpoints[i]+1...
func WildShares(poly *share.PriPoly, evalpoints []int) []*PointShare {
	shares := make([]*PointShare, 0, len(evalpoints))
	for _, ep := range evalpoints {
		ps := new(PointShare)
		ps.E = CurrentState.suite.G2().Scalar().SetInt64(int64(ep) + 1)
		ps.P = CurrentState.suite.G2().Point().Mul(poly.Eval(ep).V, nil)
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

func PruneDupShares(in []*PointShare) (unique []*PointShare) {
	unique = []*PointShare{}
	for _, ps := range in {
		isunique := true
		for _, u := range unique {
			if u.E.Equal(ps.E) {
				isunique = false
				break
			}
		}
		if isunique {
			unique = append(unique, ps)
		}

	}
	return
}

func RecoverSecretPoint(g kyber.Group, shares []*PointShare, t int) (kyber.Point, error) {
	shares = PruneDupShares(shares)
	if len(shares) < t {
		return nil, fmt.Errorf("share: not enough shares to recover secret")
	}

	acc := g.Point().Mul(g.Scalar().Zero(), nil)
	num := g.Point()
	den := g.Scalar()
	tmp := g.Scalar()

	for i, si := range shares[:t] {
		num.Set(si.P)
		den.One()

		for j, sj := range shares {
			if i == j {
				continue
			}

			num.Mul(sj.E, num)
			tmp.Sub(sj.E, si.E)
			den.Mul(den, tmp)
		}

		acc.Add(acc, g.Point().Mul(den.Inv(den), num))
	}

	return acc, nil
}
