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
	"github.com/charmbracelet/glamour"
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

type volumeDetails struct {
	volumeID   string
	attachment *string
	size       int32
	volumeType string
	tags       map[string]*string
}

// getVolume returns all volumes
func getVolumes(cmd *cobra.Command) ([]volumeDetails, error) {
	client := middleware.MustGetEC2Client(cmd.Context())
	volumes, err := client.DescribeVolumes(cmd.Context(), &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, err
	}
	tags := viper.GetStringSlice("tags")
	var volumeSummary []volumeDetails
	for _, v := range volumes.Volumes {
		details := volumeDetails{
			volumeID:   *v.VolumeId,
			size:       *v.Size,
			volumeType: string(v.VolumeType),
			tags:       make(map[string]*string),
		}
		if len(v.Attachments) > 0 {
			details.attachment = v.Attachments[0].InstanceId
		}
		for _, t := range tags {
			details.tags[t] = nil
			for _, vt := range v.Tags {
				if *vt.Key == t {
					details.tags[t] = vt.Value
				}
			}
		}
		volumeSummary = append(volumeSummary, details)
	}
	return volumeSummary, nil
}

func printVolumesByTag(cmd *cobra.Command) error {
	volumes, err := getVolumes(cmd)
	if err != nil {
		return err
	}

	tags := viper.GetStringSlice("tags")
	md := "| Volume ID | Attachment | Type | Size |"
	for _, t := range tags {
		md += fmt.Sprintf(" %s |", t)
	}
	md += "\n| --- | --- | --- | ---: |"
	for range tags {
		md += " --- |"
	}
	md += "\n"
	for _, v := range volumes {
		attachment := "unattached"
		if v.attachment != nil {
			attachment = *v.attachment
		}
		md += fmt.Sprint("|", v.volumeID, "|", attachment, "|", v.volumeType, "|", v.size, " GiB |")
		for _, t := range viper.GetStringSlice("tags") {
			if v.tags[t] == nil {
				md += "(undefined)|"
				continue
			}
			md += fmt.Sprint(*v.tags[t], "|")
		}
		md += "\n"
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	out, err := r.Render(md)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}
