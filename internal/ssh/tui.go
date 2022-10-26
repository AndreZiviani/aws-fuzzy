package ssh

import (
	"bytes"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/fzf-wrapper/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rivo/tview"
)

type Instance struct {
	ec2types.Instance
}

func (i Instance) PrintName() string {
	return fmt.Sprintf("%s (%s)", common.GetEC2Tag(i.Tags, "Name", "<missing name>"), aws.ToString(i.PrivateIpAddress))
}

func (i Instance) PrintDetails() string {
	output := bytes.NewBufferString("")

	fmt.Fprintf(output, "Name: %s\n", common.GetEC2Tag(i.Tags, "Name", "<missing name>"))

	fmt.Fprintf(output, "InstanceId: %s\n", *i.InstanceId)

	fmt.Fprintf(output, "PrivateIp: %s\n", *i.PrivateIpAddress)

	fmt.Fprintf(output, "PrivateName: %s\n", *i.PrivateDnsName)

	fmt.Fprintf(output, "PublicName: %s\n", *i.PublicDnsName)

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

type Tui struct {
	app             *tview.Application
	flex            *tview.Flex
	input           *tview.InputField
	resourceList    *tview.List
	resourceDetails *tview.TextView
	fzf             *fzfwrapper.Wrapper
	instances       []Instance
	selected        *Instance
	instanceIdx     []int
}

type FzfData struct {
	Instances []Instance
}

func NewFzfData(instancesOutput *ec2.DescribeInstancesOutput) *FzfData {
	f := FzfData{}

	f.Instances = make([]Instance, 0)

	for _, r := range instancesOutput.Reservations {
		for _, i := range r.Instances {
			tmp := Instance{}
			tmp.Instance = i // force ec2type.Instance to be my Instance type, see embedded struct doc
			f.Instances = append(f.Instances, tmp)
		}
	}

	return &f
}

func (f FzfData) FzfInputList() []string {
	out := make([]string, 0, f.FzfInputLen())

	for _, i := range f.Instances {
		out = append(out, i.PrintDetails())
	}

	return out
}

func (f FzfData) FzfInputLen() int {
	return len(f.Instances)
}

func (t *Tui) resourceListFunc(id int, text string, secondary string, shortcut rune) {
	t.resourceDetails.SetText(
		fmt.Sprintf("%s\n", secondary),
	)
}

func (t *Tui) inputFunc(text string) {
	if text == "" {
		t.resourceList.Clear()
		last := len(t.instances) - 1
		for k, v := range t.instances {
			t.instanceIdx[last-k] = k
			t.resourceList.InsertItem(
				-t.resourceList.GetItemCount()-1,
				v.PrintName(),
				v.PrintDetails(),
				0, nil,
			)
		}
		return
	}

	t.fzf.SetPattern(text)
	results, _ := t.fzf.Fuzzy()

	t.resourceList.Clear()
	t.resourceDetails.Clear()

	last := len(results) - 1
	for k, v := range results {
		t.instanceIdx[last-k] = int(v.Item.Index())
		i := t.instances[v.Item.Index()]
		t.resourceList.InsertItem(
			-t.resourceList.GetItemCount()-1,
			tview.TranslateANSI(
				i.PrintName(),
			),
			tview.TranslateANSI(v.HighlightResult()),
			0, nil,
		)
	}

	t.resourceList.SetCurrentItem(-1)
	t.resourceList.SetOffset(0, 0)
	boldItem(t.resourceList, t.resourceList.GetCurrentItem())
}

func tui(instancesOutput *ec2.DescribeInstancesOutput) (*Instance, error) {

	t := NewTui()

	fzfInput := NewFzfData(instancesOutput)
	t.fzf.SetInput(fzfInput)
	t.instances = fzfInput.Instances
	t.instanceIdx = make([]int, len(t.instances))

	last := len(t.instances) - 1

	for k, v := range t.instances {
		t.instanceIdx[last-k] = k // reverse order since we are adding items to the beggining of the list
		t.resourceList.InsertItem(
			-t.resourceList.GetItemCount()-1,
			v.PrintName(),
			v.PrintDetails(),
			0, nil,
		)
	}

	if err := t.app.SetRoot(t.flex, true).SetFocus(t.flex).Run(); err != nil {
		panic(err)
	}

	if t.selected == nil {
		// user aborted the selection (ctrl+c?)
		return nil, fmt.Errorf("Aborting by user request\n")
	}

	return t.selected, nil
}
