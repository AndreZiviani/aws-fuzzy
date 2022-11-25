package common

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	nmtypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"os"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetNMTag(tags []nmtypes.Tag, key string, missing string) string {
	// Get tag Name
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return missing
}

func GetEC2Tag(tags []ec2types.Tag, key string, missing string) string {
	// Get tag Name
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return missing
}

func SetEC2Tag(tags []ec2types.Tag, key string, value *string) []ec2types.Tag {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			tag.Key = value
			return tags
		}
	}
	newTags := append(tags, ec2types.Tag{
		Key:   aws.String(key),
		Value: value,
	})
	return newTags
}
