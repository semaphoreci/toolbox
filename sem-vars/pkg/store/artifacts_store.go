package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/semaphoreci/toolbox/sem-vars/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-vars/pkg/utils"
)

const keysInfoDirName = ".workflow-context/"

func Put(key, value string) {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2)
	defer os.Remove(file.Name())
	file.Write([]byte(value))

	contextId := utils.GetPipelineContextHierarchy()[0]
	err = execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/"+key, flags.Force)
	utils.CheckError(err, 1)
}

func Get(key string) string {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2)
	defer os.Remove(file.Name())

	contextHierarchy := utils.GetPipelineContextHierarchy()
	for _, contextID := range contextHierarchy {
		err = execArtifactCommand(Pull, keysInfoDirName+contextID+"/"+key, file.Name(), true)
		if err == nil {
			break
		}
	}

	// If err exists after the last iteration of the for loop above, we can interpret
	// that as "key value wasnt found in any pipeline context"
	// TODO currently we cant distinguish between "we cant connect to artifact registry" and "key-file does not exist"
	if err != nil {
		if flags.Fallback != "" {
			return flags.Fallback
		}
		utils.CheckError(err, 1)
	}

	byte_key, _ := os.ReadFile(file.Name())
	return string(byte_key)
}

func Delete(key string) {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2)
	defer os.Remove(file.Name())

	contextId := utils.GetPipelineContextHierarchy()[0]
	execArtifactCommand(Yank, keysInfoDirName+contextId+"/"+key, "", true)
	//The key might be present in some of the parent pipline's context as well, but we cant delete them there, as they might be used by some other pipeline.
	//We will just mark those keys as deleted inside this pipeline's context.
	err = execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/.deleted/"+key, flags.Force)
	utils.CheckError(err, 1)
}

type ArtifactCommand string

const (
	Push ArtifactCommand = "push"
	Pull                 = "pull"
	Yank                 = "yank"
)

func execArtifactCommand(command ArtifactCommand, source, dest string, force bool) error {
	var cmd *exec.Cmd
	if command == Push || command == Pull {
		if force {
			cmd = exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source, "-d", dest, "--force")
		} else {
			cmd = exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source, "-d", dest)
		}
	} else {
		cmd = exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source)
	}
	_, err := cmd.CombinedOutput()
	// fmt.Println(string(t))
	// fmt.Println("------")
	// fmt.Println(err)
	// fmt.Println("------")
	return err
}
