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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printVolumesByTag(cmd)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

// getVolume returns all volumes
func getVolumes(cmd *cobra.Command) ([]types.Volume, error) {
	client := middleware.MustGetEC2Client(cmd.Context())
	volumes, err := client.DescribeVolumes(cmd.Context(), &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, err
	}
	return volumes.Volumes, nil
}

func printVolumesByTag(cmd *cobra.Command) error {
	volumes, err := getVolumes(cmd)
	if err != nil {
		return err
	}
	tags := viper.GetStringSlice("tags")
	type volumeDetails struct {
		volumeID string
		tags     map[string]*string
	}
	var volumeSummary []volumeDetails
	for _, v := range volumes {
		details := volumeDetails{
			volumeID: *v.VolumeId,
			tags:     make(map[string]*string),
		}
		for _, t := range tags {
			details.tags[t] = nil
			for _, vt := range v.Tags {
				if *vt.Key == t {
					details.tags[t] = vt.Value
				}
			}
		}
		fmt.Print(details.volumeID + "\t")
		for _, t := range tags {
			fmt.Print(t + ":")
			if details.tags[t] == nil {
				fmt.Print("(undefined)\t")
			} else {
				fmt.Print(*details.tags[t] + "\t")
			}
		}
		fmt.Println()
		volumeSummary = append(volumeSummary, details)
	}
	return nil
}
