// Copyright 2016-2024, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/backend/display"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/result"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/spf13/cobra"
)

func newDeploymentRunCmd() *cobra.Command {
	// Flags for remote operations.
	remoteArgs := RemoteArgs{}

	var stack string
	var suppressPermalink bool

	cmd := &cobra.Command{
		Use:   "run <operation> [url]",
		Short: "Launch a deployment job on Pulumi Cloud",
		Long:  "",
		Args:  cmdutil.RangeArgs(1, 2),
		Run: cmdutil.RunResultFunc(func(cmd *cobra.Command, args []string) result.Result {
			ctx := cmd.Context()

			operation, err := apitype.ParsePulumiOperation(args[0])
			if err != nil {
				return result.FromError(err)
			}

			var url string
			if len(args) > 1 {
				url = args[1]
			}

			display := display.Options{
				Color: cmdutil.GetGlobalColorization(),
				// we only suppress permalinks if the user passes true. the default is an empty string
				// which we pass as 'false'
				SuppressPermalink: suppressPermalink,
			}

			project, _, err := readProject()
			if err != nil && !errors.Is(err, workspace.ErrProjectNotFound) {
				return result.FromError(err)
			}

			currentBe, err := currentBackend(ctx, project, display)
			if err != nil {
				return result.FromError(err)
			}

			if !currentBe.SupportsDeployments() {
				return result.FromError(fmt.Errorf("backends of this type %q do not support deployments",
					currentBe.Name()))
			}

			s, err := requireStack(ctx, stack, stackOfferNew|stackSetCurrent, display)
			if err != nil {
				return result.FromError(err)
			}

			if errResult := validateDeploymentFlags(url, remoteArgs); errResult != nil {
				return errResult
			}

			return runDeployment(ctx, cmd, display, operation, s.Ref().FullyQualifiedName().String(), url, remoteArgs)
		}),
	}

	// Remote flags
	remoteArgs.applyFlagsForDeploymentCommand(cmd)

	cmd.PersistentFlags().BoolVar(
		&suppressPermalink, "suppress-permalink", false,
		"Suppress display of the state permalink")

	cmd.PersistentFlags().StringVarP(
		&stack, "stack", "s", "",
		"The name of the stack to operate on. Defaults to the current stack")

	return cmd
}
