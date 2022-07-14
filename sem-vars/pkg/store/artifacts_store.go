package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/semaphoreci/toolbox/sem-vars/pkg/utils"
)

func Put(key, value string) error {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2)
	defer os.Remove(file.Name())
	file.Write([]byte(value))

	currentContextId := utils.GetPipelineContextHierarchy()[0]
	execArtifactCommand(Push, file.Name(), currentContextId+"/"+key)
	return nil
}

func Get(key string) string {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2)
	defer os.Remove(file.Name())

	contextHierarchy := utils.GetPipelineContextHierarchy()
	for _, contextID := range contextHierarchy {
		err = execArtifactCommand(Pull, contextID+"/"+key, file.Name())
		if err == nil {
			break
		}
	}

	// If err exists after the last iteration of the for loop above, we can interpret
	// that as "key value wasnt found in any pipeline context"
	// TODO currently we cant distinguish between "we cant connect to artifact registry" and "key-file does not exist"
	if err != nil {
		utils.CheckError(err, 1)
	}

	byte_key, _ := os.ReadFile(file.Name())
	return string(byte_key)
}

type ArtifactCommand string

const (
	Push ArtifactCommand = "push"
	Pull                 = "pull"
)

func execArtifactCommand(command ArtifactCommand, source, dest string) error {
	cmd := exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source, "-d", dest, "--force")
	_, err := cmd.CombinedOutput()
	return err
}
