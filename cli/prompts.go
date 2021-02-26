package cli

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

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
const localkeys = "Keys"
const knownshares = "Known Key Shares"
const current = "Print current setup"
const broadcast = "Manage broadcast"
const thisname = "Set Name"
const udp = "UDP Config"

func Setup() {
	for {
		prompt := promptui.Select{
			Label: "Setup",
			Items: []string{current, thisname, udp, knownshares, localkeys, up},
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
		case udp:
			UDPConfig()
		case evalpoint:
			SetEvalPoint()
		case localkeys:
			LocalKeys()
		case up:
			return
		case knownshares:
			KnownKeyShares()
		}
	}

}

const importshare = "Import a share from a file"

func KnownKeyShares() {
	for {
		suites := make([]string, len(state.CurrentState.KnownScalarShares))
		labels := make([]string, len(suites))
		i := 0
		for su, l := range state.CurrentState.KnownScalarShares {
			labels[i] = fmt.Sprintf("%s [%v of %v]", su, len(l), l[0].T)
			suites[i] = su
			i++
		}
		prpt := promptui.Select{
			Label: "Known Secret Shares",
			Size:  len(suites) + 2,
			Items: append(labels, importshare, up),
		}

		i, res, _ := prpt.Run()
		if res == up {
			return
		}
		if res == importshare {
			ImportNewShare()
		}

	}

}

func ImportNewShare() {
	prompt := promptui.Prompt{
		Label: "Sharefile name?",
	}
	filename, err := prompt.Run()
	if err != nil {
		return
	}
	err = state.CurrentState.ImportShareFile(filename)
	return
}

const samplef = "Sample incoming frames (3s)"
const viewsamples = "Viev sample frames"

var sampleTime = 3

func UDPConfig() {
	for {
		prompt := promptui.Select{
			Label: "UDP",
			Items: []string{broadcast, samplef, viewsamples, up},
		}
		_, res, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		switch res {
		case up:
			return
		case broadcast:
			Broadcast()
		case samplef:
			peers.DoSample(time.Duration(sampleTime) * time.Second)
		case viewsamples:
			Viewsamples()
		}
	}
}

func Viewsamples() {
	for {
		addrs := make([]string, len(peers.FrameSamples))
		labels := make([]string, len(peers.FrameSamples))
		i := 0
		for a, l := range peers.FrameSamples {
			labels[i] = fmt.Sprintf("%s [%v]", a, len(l))
			addrs[i] = a
			i++
		}
		prpt := promptui.Select{
			Label: "Sample UDP Frames",
			Size:  len(addrs) + 2,
			Items: append(labels, up),
		}

		i, res, _ := prpt.Run()
		if res == up {
			return
		}
		ShowFramesFrom(addrs[i])
	}

}

func ShowFramesFrom(adr string) {
	frames := make([]string, len(peers.FrameSamples[adr]))
	for i, f := range peers.FrameSamples[adr] {
		frames[i] = f.Timestamp.String()
	}
	for {
		prompt := promptui.Select{
			Label: "Frames from " + adr,
			Items: append(frames, up),
			Size:  len(frames) + 2,
		}

		i, r, _ := prompt.Run()
		if r == up {
			return
		}
		f := peers.FrameSamples[adr][i]
		fb, _ := json.MarshalIndent(f, " ", " ")
		fmt.Println(string(fb))
		err := state.CurrentState.VerifyFrame(&f)
		if err == nil {
			fmt.Println("Signature Ok")
		} else {
			fmt.Println(err)
		}
		bufio.NewReader(os.Stdin).ReadBytes('\n')

	}
}

const disablebrd = "Stop broacasting"
const enablebrd = "Start broadcasting"
const brdinterval = "Broadcast interval"

func Broadcast() {
	for {
		prompt := promptui.Select{
			Label: fmt.Sprintf("Currently broadcasting: %v", !state.CurrentState.DisableBroadcast),
			Items: []string{enablebrd, disablebrd, brdinterval, up},
		}
		_, res, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		switch res {
		case up:
			return
		case disablebrd:
			state.CurrentState.DisableBroadcast = true
			peers.S.DisableBroadcast = true
		case enablebrd:

			state.CurrentState.DisableBroadcast = false
			peers.S.DisableBroadcast = false
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
const importpriv = "Import from keyfile"
const setrandpriv = "Select Random"
const exportpriv = "Export to keyfile"

func LocalKeys() {
	for {
		prompt := promptui.Select{
			Label: "Private Key",
			Items: []string{printpriv, importpriv, setrandpriv, pubkey, inputpriv, up},
			Size:  7,
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
		case pubkey:
			if state.CurrentState.ThisPublicKey == nil {
				fmt.Println("Public key not set")
				continue
			}
			b, e := state.CurrentState.ThisPublicKey.MarshalBinary()
			if e != nil {
				fmt.Println(e)
			} else {
				fmt.Println(hex.EncodeToString(b))
			}
		case setrandpriv:
			state.SetRandomKey()
		case importpriv:
			err := ImportKeyFile()
			if err != nil {
				fmt.Println(err)
			}
		case inputpriv:
			inputPrivKeyHEX()
		}
	}
}

func inputPrivKeyHEX() {
	var deft = "01"
	if state.CurrentState.ThisSecretValue != nil {
		deft = state.CurrentState.ThisSecretValue.String()
	}
	prpt := promptui.Prompt{
		Label:    "Enter private key as hex",
		Validate: ValidateEvalPoint,
		Default:  deft,
	}
	result, err := prpt.Run()
	if err != nil {
		fmt.Println(err)
	} else {
		bi := big.NewInt(0)
		bi.SetString(result, 16)
		state.CurrentState.SetPrivKeyBytes(bi.Bytes())
	}

}

func ImportKeyFile() (err error) {
	prompt := promptui.Prompt{
		Label: "Keyfile name?",
	}
	filename, err := prompt.Run()
	if err != nil {
		return
	}
	err = state.CurrentState.ImportKeyFile(filename)
	return
}

func ValidateEvalPoint(in string) error {
	bi := big.NewInt(0)
	bi, ok := bi.SetString(in, 16)
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
		fmt.Println(err)
	} else {
		bi := big.NewInt(0)
		bi.SetString(result, 16)
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
