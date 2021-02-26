package state

import (
	"encoding/hex"
	"fmt"

	"github.com/dedis/kyber/sign/bls"
	"go.dedis.ch/kyber/v3"
)

func (cs *State) Sign(message []byte) ([]byte, error) {
	if cs.ThisSecretValue == nil {
		return nil, fmt.Errorf("private key not set")
	}
	return bls.Sign(cs.suite, cs.ThisSecretValue, message)
}

func (cs *State) VerifySignature(pkey kyber.Point, message []byte, signature []byte) error {
	return bls.Verify(cs.suite, pkey, message, signature)

}

func (cs *State) SignFrame(f *Frame) (err error) {
	if cs.ThisPublicKey == nil || cs.ThisSecretValue == nil {
		return fmt.Errorf("Keys not set")
	}
	b, err := f.FormToSign()
	sig, err := cs.Sign(b)
	f.Signature = hex.EncodeToString(sig)
	pb, err := cs.ThisPublicKey.MarshalBinary()
	f.PubKey = hex.EncodeToString(pb)
	return nil
}

func (cs *State) VerifyFrame(f *Frame) (err error) {
	if len(f.Signature) == 0 || len(f.PubKey) == 0 {
		return fmt.Errorf("No signature or public key")
	}
	pubbytes, err := hex.DecodeString(f.PubKey)
	if err != nil {
		return
	}
	pb := cs.suite.G2().Point()
	err = pb.UnmarshalBinary(pubbytes)
	if err != nil {
		return
	}
	signbytes, err := hex.DecodeString(f.Signature)
	if err != nil {
		return
	}
	msg, err := f.FormToSign()
	if err != nil {
		return
	}
	err = cs.VerifySignature(pb, msg, signbytes)
	return
}
