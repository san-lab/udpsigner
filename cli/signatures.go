package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/san-lab/udpsigner/state"
)

const splain = "Set Plaintext"
const sigmake = "Sign"
const sigverify = "Verify Signature"

var lst = new(locSigTest)

func TestLocalSignature() {

	prpt := promptui.Select{
		Label: "Test BLS signatures locally",
		Items: []string{splain, sigmake, sigverify, up},
	}

	for {
		_, result, err := prpt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		switch result {
		case splain:
			SetPlaintext(lst)
		case sigmake:
			lst.signature, err = state.CurrentState.Sign(lst.plaintext)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Signature:", hex.EncodeToString(lst.signature))
			}
		case sigverify:
			err := state.CurrentState.VerifySignature(state.CurrentState.ThisPublicKey, lst.plaintext, lst.signature)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Message:", string(lst.plaintext))
			fmt.Println("Signature:", hex.EncodeToString(lst.signature))
			if err == nil {
				fmt.Println("Signature OK")
			} else {
				fmt.Println("Error:", err)
			}
		case up:
			return

		}
	}

}

type locSigTest struct {
	plaintext []byte
	signature []byte
}

func SetPlaintext(lst *locSigTest) {

	prpt := promptui.Prompt{
		Label:   "Enter plaintext",
		Default: string(lst.plaintext),
	}
	result, err := prpt.Run()
	if err != nil {
		fmt.Println(err)
	} else {
		lst.plaintext = []byte(result)
	}

}
