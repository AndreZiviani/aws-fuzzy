package ssh

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/fzf-wrapper/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tui struct {
	app             *tview.Application
	flex            *tview.Flex
	input           *tview.InputField
	resourceList    *tview.List
	resourceDetails *tview.TextView
	fzf             *fzfwrapper.Wrapper
	instances       []ec2types.Instance
	selected        *ec2types.Instance
	instanceIdx     []int
}

// removes only the color customization at the
// beging of the string, if exists
// does NOT remove other customizations
func removeLineColor(list *tview.List, id int) {
	currentText, currentSecondary := list.GetItemText(id)
	re := regexp.MustCompile(`^\[[a-zA-Z0-9:-]+\]`)
	tmp := re.ReplaceAllString(currentText, "${1}")
	list.SetItemText(id, tmp, currentSecondary)

}
func boldItem(list *tview.List, id int) {
	if list.GetItemCount() == 0 {
		return
	}
	currentText, currentSecondary := list.GetItemText(id)
	list.SetItemText(id, fmt.Sprintf("[::b]%s", currentText), currentSecondary)
}

func printDetails(i ec2types.Instance) string {
	output := bytes.NewBufferString("")

	fmt.Fprintf(output, "Name: %s\n", common.GetEC2Tag(i.Tags, "Name", "<missing name>"))

	fmt.Fprintf(output, "InstanceId: %s\n", *i.InstanceId)

	fmt.Fprintf(output, "PrivateIp: %s\n", *i.PrivateIpAddress)

	fmt.Fprintf(output, "Subnet: %s\n", *i.SubnetId)

	fmt.Fprintf(output, "InstanceType: %s\n", i.InstanceType)

	key := "<unknown>"
	if i.KeyName != nil {
		key = *i.KeyName
	}
	fmt.Fprintf(output, "Key: %s\n", key)

	lifecycle := "OnDemand"
	if i.InstanceLifecycle != "" {
		lifecycle = string(i.InstanceLifecycle)
	}
	fmt.Fprintf(output, "Lifecycle: %s\n", lifecycle)

	fmt.Fprintf(output, "Tags:\n")
	for _, t := range i.Tags {
		fmt.Fprintf(output, "  - %s: %s\n", *t.Key, *t.Value)
	}

	fmt.Fprintf(output, "AMI: %s\n", *i.ImageId)

	fmt.Fprintf(output, "BlockDevices:\n")
	for _, b := range i.BlockDeviceMappings {
		if b.Ebs.VolumeId == nil {
			continue
		}
		fmt.Fprintf(output, "  - Id: %s\n", *b.Ebs.VolumeId)
	}

	fmt.Fprintf(output, "Interfaces:\n")
	for _, net := range i.NetworkInterfaces {
		var interfaceid, ip, subnet, vpc string

		if net.NetworkInterfaceId == nil {
			interfaceid = "<unknown>"
		} else {
			interfaceid = *net.NetworkInterfaceId
		}

		if net.PrivateIpAddress == nil {
			ip = "<unknown>"
		} else {
			ip = *net.PrivateIpAddress
		}

		if net.SubnetId == nil {
			subnet = "<unknown>"
		} else {
			subnet = *net.SubnetId
		}

		if net.VpcId == nil {
			vpc = "<classic>"
		} else {
			vpc = *net.VpcId
		}

		fmt.Fprintf(output, "  - Id: %s\n    IP: %s\n    Subnet: %s\n    Vpc: %s\n", interfaceid, ip, subnet, vpc)
	}

	fmt.Fprintf(output, "SGs:\n")
	for _, s := range i.SecurityGroups {
		var groupid, groupname string

		if s.GroupId == nil {
			groupid = "<unknown>"
		} else {
			groupid = *s.GroupId
		}

		if s.GroupName == nil {
			groupname = "<unknown>"
		} else {
			groupname = *s.GroupName
		}

		fmt.Fprintf(output, "  - %s %s\n", groupid, groupname)
	}

	vpc := "<classic>"
	if i.VpcId != nil {
		vpc = *i.VpcId
	}
	fmt.Fprintf(output, "Vpc: %s\n", vpc)

	return output.String()
}

func NewTui() *Tui {
	t := Tui{
		app: tview.NewApplication(),
		resourceDetails: tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true),
		resourceList: tview.NewList().
			ShowSecondaryText(false).
			SetSelectedBackgroundColor(tcell.ColorDarkSlateGray).
			SetSelectedTextColor(tcell.ColorWhite).
			SetMainTextColor(tcell.ColorDarkGray),
		input: tview.NewInputField().
			SetLabel(">: "),
		flex: tview.NewFlex(),
		fzf:  fzfwrapper.NewWrapper(fzfwrapper.WithSortBy(fzfwrapper.ByScore, fzfwrapper.ByPosition, fzfwrapper.ByLength)),
	}

	t.app.EnableMouse(true)
	t.resourceDetails.SetBorder(true)
	t.resourceList.SetBorder(true)

	t.resourceList.SetChangedFunc(func(id int, text string, secondary string, shortcut rune) {
		t.resourceDetails.SetText(
			fmt.Sprintf("%s\n", secondary),
		)
	})

	t.input.SetChangedFunc(func(text string) {
		idx := make([]int, 0)
		if text == "" {
			t.resourceList.Clear()
			for k, v := range t.instances {
				idx = append([]int{k}, idx...) // reverse order since we are adding items to the beggining of the list
				t.resourceList.InsertItem(
					-t.resourceList.GetItemCount()-1,
					fmt.Sprintf("%s (%s)", common.GetEC2Tag(v.Tags, "Name", "<missing name>"), aws.ToString(v.PrivateIpAddress)),
					fmt.Sprintf("%s", printDetails(v)),
					0, nil,
				)
			}
			t.instanceIdx = idx
			return
		}

		t.fzf.SetPattern(text)
		results, _ := t.fzf.Fuzzy()

		t.resourceList.Clear()
		t.resourceDetails.Clear()

		for _, v := range results {
			idx = append([]int{int(v.Item.Index())}, idx...) // reverse order since we are adding items to the beggining of the list
			i := t.instances[v.Item.Index()]
			t.resourceList.InsertItem(
				-t.resourceList.GetItemCount()-1,
				tview.TranslateANSI(
					fmt.Sprintf("%s (%s)", common.GetEC2Tag(i.Tags, "Name", "<missing name>"), aws.ToString(i.PrivateIpAddress)),
				),
				tview.TranslateANSI(v.HighlightResult()),
				0, nil,
			)
		}
		t.instanceIdx = idx
		t.resourceList.SetCurrentItem(-1)
		t.resourceList.SetOffset(0, 0)
		boldItem(t.resourceList, t.resourceList.GetCurrentItem())
	})

	t.flex.SetDirection(tview.FlexRow).
		// Horizontal view, textView
		AddItem(tview.NewFlex().
			// Vertical view, options | details
			AddItem(t.resourceList, 0, 1, false).
			AddItem(t.resourceDetails, 0, 1, false),
			0, 1, false).
		// Horizontal view, input field
		AddItem(t.input, 1, 1, true)

	t.setCaptureEvents()
	return &t
}

func (t *Tui) setCaptureEvents() {
	// Capture key events to perform custom actions
	// Configure TAB key to cycle between windows
	// Configure Up/Down key in input screen to scroll the list
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		where := t.app.GetFocus()
		switch k {
		case tcell.KeyEnter:
			current := t.resourceList.GetCurrentItem() // current index selected from list
			instanceIdx := t.instanceIdx[current]      // offset of instances list
			t.selected = &t.instances[instanceIdx]
			t.app.Stop()
			return nil
		case tcell.KeyTab:
			switch where {
			case t.resourceDetails:
				// next window
				t.app.SetFocus(t.input)
				return nil
			case t.resourceList:
				// next window
				t.app.SetFocus(t.resourceDetails)
				return nil
			case t.input:
				// next window
				t.app.SetFocus(t.resourceList)
				return nil
			}
		case tcell.KeyBacktab:
			switch where {
			case t.resourceDetails:
				// previous window
				t.app.SetFocus(t.resourceList)
				return nil
			case t.resourceList:
				// previous window
				t.app.SetFocus(t.input)
				return nil
			case t.input:
				// previous window
				t.app.SetFocus(t.resourceDetails)
				return nil
			}
		case tcell.KeyUp:
			switch where {
			case t.input, t.resourceList:
				// list up
				current := t.resourceList.GetCurrentItem()
				previous := current - 1
				if previous < 0 {
					previous = t.resourceList.GetItemCount() - 1
				}
				removeLineColor(t.resourceList, current)
				boldItem(t.resourceList, previous)
				t.resourceList.SetCurrentItem(previous)
				return nil
			}
		case tcell.KeyDown:
			switch where {
			case t.input, t.resourceList:
				// list down
				current := t.resourceList.GetCurrentItem()
				next := (current + 1) % t.resourceList.GetItemCount()
				removeLineColor(t.resourceList, current)
				boldItem(t.resourceList, next)
				t.resourceList.SetCurrentItem(next)
				return nil
			}
		}
		return event
	})
}

func tui(instancesOutput *ec2.DescribeInstancesOutput) (*ec2types.Instance, error) {

	t := NewTui()

	inputData := make([]string, 0)
	instances := make([]ec2types.Instance, 0)
	idx := make([]int, 0)

	for _, r := range instancesOutput.Reservations {
		for _, i := range r.Instances {
			inputData = append(inputData, printDetails(i))
			instances = append(instances, i)
		}
	}

	t.fzf.SetInput(inputData)
	t.instances = instances

	for k, v := range instances {
		idx = append([]int{k}, idx...) // reverse order since we are adding items to the beggining of the list
		t.resourceList.InsertItem(
			-t.resourceList.GetItemCount()-1,
			fmt.Sprintf("%s (%s)", common.GetEC2Tag(v.Tags, "Name", "<missing name>"), aws.ToString(v.PrivateIpAddress)),
			fmt.Sprintf("%s", printDetails(v)),
			0, nil,
		)
	}

	t.instanceIdx = idx

	if err := t.app.SetRoot(t.flex, true).SetFocus(t.flex).Run(); err != nil {
		panic(err)
	}

	if t.selected == nil {
		// user aborted the selection (ctrl+c?)
		return nil, fmt.Errorf("Aborting by user request\n")
	}

	return t.selected, nil
}
