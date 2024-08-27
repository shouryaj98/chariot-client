package model

import "time"

type Capability interface {
	Match() bool
	Invoke() error
	Secret() string
	Timeout() int
}

type Integration interface {
	Push(risk Risk)
}

type History struct {
	From    string `json:"from"`
	To      string `json:"to"`
	By      string `json:"by"`
	Comment string `json:"comment"`
	Updated string `json:"updated"`
}

const (
	// shared:status
	Deleted string = "D"
	// asset:status
	Active     string = "A"
	ActiveLow  string = "AL"
	ActiveHigh string = "AH"
	Frozen     string = "F"
	FrozenLow  string = "FL"
	FrozenHigh string = "FH"
	// job:status
	Queued  string = "JQ"
	Running string = "JR"
	Fail    string = "JF"
	Pass    string = "JP"
	// risk:status
	Triage                 string = "T"
	TriageInfo             string = "TI"
	TriageLow              string = "TL"
	TriageMedium           string = "TM"
	TriageHigh             string = "TH"
	TriageCritical         string = "TC"
	Open                   string = "O"
	OpenInfo               string = "OI"
	OpenLow                string = "OL"
	OpenMedium             string = "OM"
	OpenHigh               string = "OH"
	OpenCritical           string = "OC"
	Remediated             string = "R"
	RemediatedInfo         string = "RI"
	RemediatedLow          string = "RL"
	RemediatedMedium       string = "RM"
	RemediatedHigh         string = "RH"
	RemediatedCritical     string = "RC"
	MachineOpen            string = "MO"
	MachineOpenInfo        string = "MOI"
	MachineOpenLow         string = "MOL"
	MachineOpenMedium      string = "MOM"
	MachineOpenHigh        string = "MOH"
	MachineOpenCritical    string = "MOC"
	MachineDeleted         string = "MD"
	MachineDeletedInfo     string = "MDI"
	MachineDeletedLow      string = "MDL"
	MachineDeletedMedium   string = "MDM"
	MachineDeletedHigh     string = "MDH"
	MachineDeletedCritical string = "MDC"
	// asset:source
	Discovered string = "discovered"
	Provided   string = "provided"
)

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func Future(hours int) int64 {
	return time.Now().UTC().Add(time.Duration(hours) * time.Hour).Unix()
}
