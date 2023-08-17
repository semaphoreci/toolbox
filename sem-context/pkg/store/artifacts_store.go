package store

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
)

const keysInfoDirName = ".workflow-context/"

type ArtifactStore struct{}

func (_ *ArtifactStore) Put(key, value, contextId string) error {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return &utils.Error{ErrorMessage: "Cant create temp file to store contents from artifacts", ExitCode: 2}
	}
	defer os.Remove(file.Name())

	_, err = file.Write([]byte(value))
	if err != nil {
		return &utils.Error{ErrorMessage: fmt.Sprintf("error writing value to file: %v", err), ExitCode: 2}
	}

	artifact_output, err := execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/"+key)
	if err != nil {
		log.New(os.Stderr, "", 0).Println(artifact_output)
		return &utils.Error{ErrorMessage: "Cant execute artifacts push command to store key-value pair", ExitCode: 2}
	}

	//Since the key is stored, delete it from '.deleted' dir, in case it was marked as deleted before
	if output, err := execArtifactCommand(Yank, keysInfoDirName+contextId+"/.deleted/"+key, ""); err != nil {
		log.Printf("error executing artifact command: %v. Output: %s\n", err, output)
	}

	return nil
}

func (_ *ArtifactStore) Get(key, contextId string) (string, error) {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", &utils.Error{ErrorMessage: "Cant create temp file to store contents from artifacts", ExitCode: 2}
	}
	defer os.Remove(file.Name())

	artifact_output, err := execArtifactCommand(Pull, keysInfoDirName+contextId+"/"+key, file.Name())
	if err != nil {
		// Since 'artifact' CLI always returns 1, this is the only way to check if
		// communication with artifact server is the problem, of key just does not exist
		if strings.Contains(artifact_output, "404 status code") {
			return "", &utils.Error{ErrorMessage: fmt.Sprintf("Cant find the key '%s'", key), ExitCode: 1}
		} else {
			log.New(os.Stderr, "", 0).Panicln(artifact_output)
			return "", &utils.Error{ErrorMessage: "Error with establishing connection with artifact server", ExitCode: 2}
		}
	}

	byte_key, _ := os.ReadFile(file.Name())
	return string(byte_key), nil
}

func (_ *ArtifactStore) Delete(key, contextId string) error {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return &utils.Error{
			ErrorMessage: "Cant create temp file needed to perform delete operation when using artifacts as store",
			ExitCode:     2,
		}
	}
	defer os.Remove(file.Name())

	if output, err := execArtifactCommand(Yank, keysInfoDirName+contextId+"/"+key, ""); err != nil {
		log.Printf("error executing artifact command: %v. Output: %s\n", err, output)
	}

	// The key might be present in some of the parent pipline's context as well, but we cant delete them there, as they might be used by some other pipeline.
	// We will just mark those keys as deleted inside this pipeline's context.
	artifact_output, err := execArtifactCommand(Push, file.Name(), keysInfoDirName+contextId+"/.deleted/"+key)
	if err != nil {
		// Since 'artifact' CLI always returns 1, this is the only way to check if
		// communication with artifact server is the problem, of key just does not exist
		if !strings.Contains(artifact_output, "Artifact not found") {
			log.New(os.Stderr, "", 0).Panicln(artifact_output)
			return &utils.Error{ErrorMessage: "Error with establishing connection with artifact server", ExitCode: 2}
		}
	}
	return nil
}

type ArtifactCommand string

const (
	Push ArtifactCommand = "push"
	Pull                 = "pull"
	Yank                 = "yank"
)

func (_ *ArtifactStore) CheckIfKeyDeleted(key, contextId string) (bool, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return false, &utils.Error{
			ErrorMessage: "Cant create temp file needed which is needed when using artifacts as store",
			ExitCode:     2,
		}
	}
	defer os.RemoveAll(dir)

	if output, err := execArtifactCommand(Pull, keysInfoDirName+contextId+"/.deleted/", dir); err != nil {
		log.Printf("error executing artifact command: %v. Output: %s\n", err, output)
	}

	all_deleted_key_files, _ := ioutil.ReadDir(dir)
	for _, deleted_key_file := range all_deleted_key_files {
		if key == deleted_key_file.Name() {
			return true, nil
		}
	}
	return false, nil
}

func execArtifactCommand(command ArtifactCommand, source, dest string) (string, error) {
	var cmd *exec.Cmd
	if command == Push || command == Pull {
		// #nosec
		cmd = exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source, "-d", dest, "--force")
	} else {
		// #nosec
		cmd = exec.Command("artifact", fmt.Sprintf("%v", command), "workflow", source)
	}
	artifact_output, err := cmd.CombinedOutput()
	return string(artifact_output), err
}
