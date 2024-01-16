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

package middleware

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

// NewEC2Client is middleware to use for CLI commands that require a Super
// API client. It creates a new Super API client and adds it to the provided
// command's context. If a Consul client does not exist in the context, it will
// be created and added to the context.
func NewEC2Client(cmd *cobra.Command, _ []string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	client := ec2.NewFromConfig(cfg)
	cmd.SetContext(context.WithValue(cmd.Context(), EC2ClientContext, client))
	return nil
}

// MustGetEC2Client returns the Django EC2 from the context. If the client
// does not exist in the context, this function panics.
func MustGetEC2Client(ctx context.Context) *ec2.Client {
	c := ctx.Value(EC2ClientContext)
	if c == nil {
		panic("EC2 client not set in context")
	}
	return c.(*ec2.Client)
}
