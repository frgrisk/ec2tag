/*
Copyright © 2024 FRG

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/frgrisk/ec2tag/cmd/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncTags(cmd)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func syncTags(cmd *cobra.Command) error {
	volumes, err := getVolumes(cmd)
	if err != nil {
		return err
	}
	// TODO: add force flag
	tags := viper.GetStringSlice("tags")
	instanceMap, err := getInstances(cmd)
	if err != nil {
		return err
	}
	for _, v := range volumes {
		for _, t := range tags {
			if v.tags[t] == nil {
				if v.attachment == nil {
					fmt.Println("Volume", v.volumeID, "is not attached to an instance")
					continue
				}
				tagValue := instanceMap[*v.attachment].tags[t]
				if tagValue != nil {
					fmt.Printf("Adding tag %s:%s to volume %s\n", t, *tagValue, v.volumeID)
					// TODO: do this in bulk
					err = tagResource(cmd, v.volumeID, t, *tagValue)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func tagResource(cmd *cobra.Command, resourceID string, tagKey string, tagValue string) error {
	client := middleware.MustGetEC2Client(cmd.Context())
	_, err := client.CreateTags(cmd.Context(), &ec2.CreateTagsInput{
		Resources: []string{resourceID},
		Tags: []types.Tag{
			{
				Key:   &tagKey,
				Value: &tagValue,
			},
		},
	})
	return err
}

type InstanceDetails struct {
	tags map[string]*string
}

func getInstances(cmd *cobra.Command) (map[string]InstanceDetails, error) {
	client := middleware.MustGetEC2Client(cmd.Context())
	instances, err := client.DescribeInstances(cmd.Context(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}
	instanceMap := make(map[string]InstanceDetails)
	tags := viper.GetStringSlice("tags")
	for _, r := range instances.Reservations {
		for _, i := range r.Instances {
			instanceMap[*i.InstanceId] = InstanceDetails{
				tags: make(map[string]*string),
			}
			for _, t := range tags {
				instanceMap[*i.InstanceId].tags[t] = nil
				for _, it := range i.Tags {
					if *it.Key == t {
						instanceMap[*i.InstanceId].tags[t] = it.Value
					}
				}
			}
		}
	}
	return instanceMap, nil
}
