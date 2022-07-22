package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
)

const keysInfoDirName = ".workflow-context/"

func Put(key, value string) {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2, "Cant create temp file to store contents from artifacts")
	defer os.Remove(file.Name())
	file.Write([]byte(value))

	contextId := utils.GetPipelineContextHierarchy()[0]
	artifact_output, err := execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/"+key, flags.Force)
	utils.CheckError(err, 1, artifact_output)

	//Since the key is stored, delete it from '.deleted' dir, in case it was marked as deleted before
	execArtifactCommand(Yank, keysInfoDirName+contextId+"/.deleted/"+key, "", true)
	fmt.Fprintf(os.Stdout, "Key-value pair successfully stored")
}

func Get(key string) string {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2, "Cant create temp file from which key-value pair will be uploaded to artifacts")
	defer os.Remove(file.Name())

	contextHierarchy := utils.GetPipelineContextHierarchy()
	found_the_key := false
	for _, contextID := range contextHierarchy {
		_, err = execArtifactCommand(Pull, keysInfoDirName+contextID+"/"+key, file.Name(), true)
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
			fmt.Fprintf(os.Stdout, "Could not find key '%s', using the fallback value", key)
			return flags.Fallback
		}
		utils.CheckError(err, 1, fmt.Sprintf("Cant find the key '%s'", key))
	}

	byte_key, _ := os.ReadFile(file.Name())
	return string(byte_key)
}

func Delete(key string) {
	file, err := ioutil.TempFile("", "")
	utils.CheckError(err, 2, "Can't create a temporary file which is needed to performe 'delete' operation")
	defer os.Remove(file.Name())

	contextId := utils.GetPipelineContextHierarchy()[0]
	execArtifactCommand(Yank, keysInfoDirName+contextId+"/"+key, "", true)
	//The key might be present in some of the parent pipline's context as well, but we cant delete them there, as they might be used by some other pipeline.
	//We will just mark those keys as deleted inside this pipeline's context.
	artifact_output, err := execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/.deleted/"+key, true)
	utils.CheckError(err, 1, artifact_output)
}

type ArtifactCommand string

const (
	Push ArtifactCommand = "push"
	Pull                 = "pull"
	Yank                 = "yank"
)

func checkIfKeyDeleted(contextID, key string) bool {
	dir, err := ioutil.TempDir("", "")
	utils.CheckError(err, 2, "Cant create a temporary file")
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

func execArtifactCommand(command ArtifactCommand, source, dest string, force bool) (string, error) {
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
	artifact_output, err := cmd.CombinedOutput()
	return string(artifact_output), err
}
