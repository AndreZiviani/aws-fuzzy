package version

import (
	"fmt"
)

func (p *Version) Execute(args []string) error {
	if len(versionTag) > 0 {
		fmt.Printf("aws-fuzzy %s\n", versionTag)
	} else {
		fmt.Printf("Unknown version, manually compiled from git?\n")
	}
	return nil
}
