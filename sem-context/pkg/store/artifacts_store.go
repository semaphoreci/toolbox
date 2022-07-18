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
	found_the_key := false
	for _, contextID := range contextHierarchy {
		err = execArtifactCommand(Pull, keysInfoDirName+contextID+"/"+key, file.Name(), true)
		if err == nil {
			found_the_key = true
			break
		}
		//If key is deleted, we dont need to go looking for it in parent contexts
		key_deleted := checkIfKeyDeleted(contextID, key)
		if key_deleted {
			break
		}
	}

	if !found_the_key {
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

func checkIfKeyDeleted(contextID, key string) bool {
	dir, err := ioutil.TempDir("", "")
	utils.CheckError(err, 2)
	defer os.RemoveAll(dir)

	execArtifactCommand(Pull, keysInfoDirName+contextID+"/.deleted/", dir, true)

	all_deleted_key_files, _ := ioutil.ReadDir(dir)
	for _, deleted_key_file := range all_deleted_key_files {
		if key == deleted_key_file.Name() {
			return true
		}
	}
	return false
}

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
	return err
}
