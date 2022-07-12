package store

import (
	"log"
	"os"

	errutil "github.com/semaphoreci/artifact/pkg/errors"
	artifact_files "github.com/semaphoreci/artifact/pkg/files"
	artifact_hub "github.com/semaphoreci/artifact/pkg/hub"
	artifact_storage "github.com/semaphoreci/artifact/pkg/storage"
)

func Put(key, value string) error {
	hubClient, err := artifact_hub.NewClient()
	errutil.Check(err)
	// By passing empty string to resolver we are making sure SEMAPHORE_PROJECT_ID env variable is set, no fallback options
	// TODO I assume resourceType should be workflow (or maybe entire project?)
	resolver, err := artifact_files.NewPathResolver(artifact_files.ResourceTypeProject, "")
	errutil.Check(err)
	err = createLocalFile(key, value)
	if err != nil {
		errutil.Check(err)
		return err
	}
	pushOptions := artifact_storage.PushOptions{
		SourcePath:          key,
		DestinationOverride: "",
		Force:               true,
	}
	_, err = artifact_storage.Push(hubClient, resolver, pushOptions)
	deleteLocalFile(key)
	if err != nil {
		return err
	}
	return nil
}

func Get(key string) string {
	hubClient, err := artifact_hub.NewClient()
	errutil.Check(err)
	resolver, err := artifact_files.NewPathResolver(artifact_files.ResourceTypeProject, "")
	errutil.Check(err)
	pullOptions := artifact_storage.PullOptions{
		SourcePath:          key,
		DestinationOverride: key,
		Force:               true,
	}
	_, err = artifact_storage.Pull(hubClient, resolver, pullOptions)
	errutil.Check(err)
	value, err := readFileContents(key)
	errutil.Check(err)
	deleteLocalFile(key)
	return value
}

func createLocalFile(file_name, file_contents string) error {
	err := os.WriteFile(file_name, []byte(file_contents), 0666)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func readFileContents(file_name string) (string, error) {
	data, err := os.ReadFile(file_name)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(data[:]), nil
}

func deleteLocalFile(file_name string) {
	os.Remove(file_name)
}
