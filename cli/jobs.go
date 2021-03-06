package cli

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/san-lab/udpsigner/state"
)

func Jobs() {

	for {
		prompt := promptui.Select{
			Label: "Select Action ",
			Items: []string{joblist, donelist, newjob, up},
			//AddLabel: endpoint,
			Size: 7,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case joblist:
			PendingJobs()
		case donelist:
			DoneJobs()

		case newjob:
			NewJob()

		case up:
			return
		}

	}
}

func DoneJobs() {
	for {
		jobs := []string{}
		ids := []string{}
		for i, jb := range state.CurrentState.DoneJobs {
			item := fmt.Sprintf("%v form %v", i, jb.AgentID)
			jobs = append(jobs, item)
			ids = append(ids, i)
		}
		items := append(jobs, up)
		prompt := promptui.Select{
			Label: "Finished Jobs",
			Items: items,
			//AddLabel: endpoint,
			Size: len(items),
		}

		i, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case up:
			return
		}

		if i < len(ids) {
			j := state.CurrentState.DoneJobs[ids[i]]
			fmt.Println("Job Type:", j.Type)
			fmt.Println("Finished:", j.FinishedAt)
			fmt.Println("Result:", j.FinalResult)
		}

	}
}

func PendingJobs() {
	for {
		jobs := []string{}
		ids := []string{}
		for i, jb := range state.CurrentState.PendingJobs {
			item := fmt.Sprintf("%v form %v", i, jb.AgentID)
			jobs = append(jobs, item)
			ids = append(ids, i)
		}
		items := append(jobs, up)
		prompt := promptui.Select{
			Label: "Pending Jobs",
			Items: items,
			//AddLabel: endpoint,
		}

		i, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case up:
			return
		}
		j := state.CurrentState.PendingJobs[ids[i]]
		if j != nil {

			ManageJob(j)
		}

	}
}

const japprove = "Approve"
const jdelete = "Delete"
const jresend = "Send again"
const jdetails = "Details"

func ManageJob(jb *state.Job) {
	label := fmt.Sprintf("Job %v from %v accepted %v", jb.ID, jb.AgentID, jb.Accepted)
	items := []string{jdetails, japprove, jresend, jdelete, up}
	for {
		prompt := promptui.Select{
			Label: label,
			Items: items,
			//AddLabel: endpoint,
			Size: 5,
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case jdetails:
			JobDetails(jb)
		case jdelete:
			DeleteJob(jb)
			return
		case japprove:
			state.CurrentState.ProcessJob(jb)
			return
		case jresend:
			if !jb.Finished {
				state.JobToBroadcastQueue(jb, 1)
			} else {
				fmt.Println("Job already marked as finished")
			}

		case up:
			return

		}

	}
}

func JobDetails(jb *state.Job) {

	fmt.Println(jb.JobDetailsString())

}

func DeleteJob(jb *state.Job) {
	delete(state.CurrentState.PendingJobs, jb.ID)

}

const testjob = "Test task"
const mpsignjob = "New M-P signature"
const pubkeyjob = "Public Key assembly"
const localsign = "Local Signature"

func NewJob() {

	prompt := promptui.Select{
		Label: "Select Action ",
		Items: []string{testjob, pubkeyjob, mpsignjob, localsign, up},
		//AddLabel: endpoint,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {
	case testjob:
		NewTestJob()
	case pubkeyjob:
		state.NewMPPubJobStart()
	case mpsignjob:
		state.NewMPSignJobStart()
	case localsign:
		TestLocalSignature()

	case up:
		return
	}

}

func NewTestJob() {
	prpt := promptui.Prompt{
		Label:   "Set the test payload",
		Default: "Are you there?",
	}
	result, err := prpt.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	j := state.CurrentState.NewTestJob(result)
	//state.CurrentState.ProcessTestJob(j)
	state.JobToBroadcastQueue(j, 1)

}

func BroadcastQueue() {
	for {
		queue := []string{}
		for id, n := range state.CurrentState.JobBroadcast {
			queue = append(queue, fmt.Sprintf("%v(%v)", id, n))
		}
		items := append(queue, up)
		prompt := promptui.Select{
			Label: "Pending Jobs",
			Items: items,
			//AddLabel: endpoint,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case up:
			return
		}

	}
}
