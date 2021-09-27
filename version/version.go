package version

import "fmt"

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	date                 = ""
	VersionInfo          *info
)

type info struct {
	Application   string `json:"Application"`
	VersionString string `json:"VersionString"`
	Commit        string `json:"Commit"`
	Date          string `json:"Date"`
}

func newInfo(application string) *info {
	return &info{
		Application:   application,
		VersionString: versionString,
		Commit:        commit,
		Date:          date,
	}
}

func (i *info) String() string {
	return fmt.Sprintf("%s Version: %s Commit: %s Date: %s", i.Application, i.VersionString, i.Commit, i.Date)
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
	VersionInfo = newInfo("arduino-cloud-cli")
}
