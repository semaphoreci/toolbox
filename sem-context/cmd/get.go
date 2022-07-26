package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
	"github.com/semaphoreci/toolbox/sem-context/pkg/validators"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a variable",
	Run:   RunGetCmd,
}

func RunGetCmd(cmd *cobra.Command, args []string) {
	utils.CheckError(validators.ValidateGetAndDeleteArguments(args))
	key := args[0]

	value, err := SearchForKeyInAllContexts(key)
	if value != "" {
		fmt.Println(value)
		return
	}
	if err != nil && err.(*utils.Error).ExitCode == 1 && flags.Fallback != "" {
		fmt.Println(flags.Fallback)
		return
	}
	utils.CheckError(err)
}

// Goes from current context all the way to the root context (context<=>pipeline) and
// searches for given key.
func SearchForKeyInAllContexts(key string) (string, error) {
	contextHierarchy := utils.GetPipelineContextHierarchy()
	var err error
	for _, contextID := range contextHierarchy {
		value, err := store.Get(key, contextID)
		if err == nil {
			return value, nil
		}
		if err.(*utils.Error).ExitCode == 2 {
			return "", err
		}
		deleted, _ := store.CheckIfKeyDeleted(contextID, key)
		if deleted {
			return "", &utils.Error{ErrorMessage: fmt.Sprintf("Cant find the key '%s'", key), ExitCode: 1}
		}
	}
	return "", err
}

func init() {
	getCmd.Flags().StringVar(&flags.Fallback, "fallback", "", "Default value to be returned if key does not exist.")
	RootCmd.AddCommand(getCmd)
}
