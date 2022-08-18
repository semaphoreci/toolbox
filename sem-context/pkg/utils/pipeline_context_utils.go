package utils

import (
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const artifactIDregex = "SEMAPHORE_PIPELINE_(.*)_ARTEFACT_ID"

//For this tool Artifact ID's are used as Context ID's
//Returns slice with context ID's where first element of slice is current pipelines ID, and last element is ID of the first pipeline
//inside the workflow
func GetPipelineContextHierarchy() []string {
	contextIDs_map := extractArtifactIdsFromEnvVariables()

	return sortContextIDs(contextIDs_map)
}

func extractArtifactIdsFromEnvVariables() map[int]string {
	contextIDs := make(map[int]string)

	for _, element := range os.Environ() {
		env_var := strings.Split(element, "=")
		if match, _ := regexp.MatchString(artifactIDregex, env_var[0]); match {
			idOrderNum := extractPipelineOrdinalNumber(env_var[0])
			contextIDs[idOrderNum] = env_var[1]
		}
	}

	return contextIDs
}

// Env variables that contain artefact id for given pipeline contain info about what is
// the order of execution of all pipelines in the given workflow:
// SEMAPHORE_PIPELINE_1_ARTEFACT_ID=...
// SEMAPHORE_PIPELINE_13_ARTEFACT_ID...
// This function extracts ordinal number of given pipeline from env variable name.
// Artifact ID with the highest ordinal number is artifact id of current pipeline.
func extractPipelineOrdinalNumber(env_var string) int {
	re := regexp.MustCompile(artifactIDregex)
	substrings := re.FindAllStringSubmatch(env_var, 1)
	orderNum, _ := strconv.Atoi(substrings[0][1])
	return orderNum
}

func sortContextIDs(contextIDs_map map[int]string) []string {
	keys := make([]int, 0, len(contextIDs_map))
	for key := range contextIDs_map {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	contextIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		contextIDs = append(contextIDs, contextIDs_map[key])
	}
	return contextIDs
}
