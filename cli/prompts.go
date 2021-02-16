package cli

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/san-lab/udpsigner/peers"
	"github.com/san-lab/udpsigner/state"
)

const exit = "EXIT"
const setup = "Setup"
const jobs = "Jobs"
const peerslist = "Peers"
const joblist = "Pending Jobs"
const donelist = "Done Jobs"
const localjoblist = "Local Actions"

const bcastqueue = "Broadcast Queue"
const newjob = "New Job"

const up = "Back"
const refresh = "Refresh"

func Top() {

	for {
		prompt := promptui.Select{
			Label: "Select Action ",
			Items: []string{setup, peerslist, jobs, exit},
			//AddLabel: endpoint,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case setup:
			Setup()
		case peerslist:
			Peers()
		case jobs:
			Jobs()
		case exit:
			fmt.Println("Thanks for using the UdpSigner")
			return
		}

	}
}

func Peers() {

	for {
		pc := len(peers.Nodes)
		items := make([]string, pc+2)
		i := 0
		for k, v := range peers.Nodes {
			items[i] = k + "/" + string(v.ID) + ": " + v.Name + "\tLast seen: " + v.LastSeen.Format("2006-01-02 15:04:05")
			i++
		}
		items[i] = refresh
		i++
		items[i] = up
		label := "Detected peers"
		if pc == 0 {
			label = "No peers detected"
		}
		prompt := promptui.Select{
			Label: label,
			Items: items,
			Size:  pc + 3,
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch result {
		case refresh:
			continue
		case up:
			return
		}
	}
}

const evalpoint = "Evaluation point"
const pubkey = "Public key"
const privkey = "Private key"
const current = "Print current setup"
const broadcast = "Disable broadcast"
const thisname = "Set Name"

func Setup() {
	for {
		prompt := promptui.Select{
			Label: "Setup",
			Items: []string{current, thisname, evalpoint, pubkey, privkey, up},
			Size:  7,
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch result {
		case thisname:
			SetName()
		case current:
			PrintCurrentSetup()
		case evalpoint:
			SetEvalPoint()
		case pubkey:
			if state.CurrentState.ThisPublicKey == nil {
				fmt.Println("Kublic key not set")
				continue
			}
			b, e := state.CurrentState.ThisPublicKey.MarshalBinary()
			if e != nil {
				fmt.Println(e)
			} else {
				fmt.Println(hex.EncodeToString(b))
			}
		case privkey:
			PrivateKey()
		case up:
			return
		}
	}

}

func PrintCurrentSetup() {
	fmt.Println("AgentName", state.CurrentState.ThisName)
	fmt.Println("AgentID", state.CurrentState.ThisId)
	fmt.Println("PubKey", state.CurrentState.ThisPublicKey)
	fmt.Println("EvalPt", state.CurrentState.ThisEvaluationPoint.String())
	fmt.Println("Broadcast", !state.CurrentState.DisableBroadcast)
}

const printpriv = "Print as HEX"
const inputpriv = "Input as HEX"
const setrandpriv = "Select Random"
const exportpriv = "Export to keyfile"

func PrivateKey() {
	for {
		prompt := promptui.Select{
			Label: "Private Key",
			Items: []string{printpriv, inputpriv, setrandpriv, exportpriv, up},
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch result {
		case up:
			return
		case printpriv:
			fmt.Println(state.CurrentState.ThisSecretValue)
		case thisname:
			SetName()
		case setrandpriv:
			state.SetRandomKey()
		}
	}
}

func ValidateEvalPoint(in string) error {
	bi := big.NewInt(0)
	bi, ok := bi.SetString(in, 10)
	if !ok {
		return fmt.Errorf("Invalid input")
	}
	return nil
}

func SetEvalPoint() {

	prpt := promptui.Prompt{
		Label:    "Set evaluation point",
		Validate: ValidateEvalPoint,
		Default:  state.CurrentState.ThisEvaluationPoint.String(),
	}
	result, err := prpt.Run()
	if err != nil {
		//fmt.Println(err)
	} else {
		bi := big.NewInt(0)
		bi.SetString(result, 10)
		state.CurrentState.ThisEvaluationPoint.SetBytes(bi.Bytes())
	}

}

func SetName() {
	prpt := promptui.Prompt{
		Label:   "Set Agent's name",
		Default: "",
	}
	result, err := prpt.Run()
	if err != nil {
		//fmt.Println(err)
	} else {
		state.CurrentState.ThisName = result
	}

}
