package cli

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/semaphoreci/toolbox/test-results/pkg/parsers"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LoadFiles checks if path exists and can be `stat`ed at given `path`
func LoadFiles(inPaths []string, ext string) ([]string, error) {
	paths := []string{}

	for _, path := range inPaths {
		file, err := os.Stat(path)

		if err != nil {
			logger.Error("Input file read failed: %v", err)
			return paths, err
		}

		switch file.IsDir() {
		case true:
			err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
				if d.Type().IsRegular() {
					switch filepath.Ext(d.Name()) {
					case ext:
						paths = append(paths, path)
					}
				}
				return nil
			})

			if err != nil {
				logger.Error("Walking through directory %s failed %v", path, err)
				return paths, err
			}

		case false:
			switch filepath.Ext(file.Name()) {
			case ext:
				paths = append(paths, path)
			}
		}
	}

	sort.Strings(paths)

	return paths, nil
}

// CheckFile checks if path exists and can be `stat`ed at given `path`
func CheckFile(path string) (string, error) {
	_, err := os.Stat(path)

	if err != nil {
		logger.Error("Input file read failed: %v", err)
		return "", err
	}

	return path, nil
}

// FindParser finds parser according to file type or flag specified by user
func FindParser(path string, cmd *cobra.Command) (parser.Parser, error) {

	parserName, err := cmd.Flags().GetString("parser")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return nil, err
	}

	parser, err := parsers.FindParser(parserName, path)
	if err != nil {
		logger.Error("Could not find parser: %v", err)
		return nil, err
	}
	logger.Info("Using %s parser", parser.GetName())
	return parser, nil
}

// Parse parses file at `path` with given `parser`
func Parse(p parser.Parser, path string, cmd *cobra.Command) (parser.Result, error) {
	result := parser.NewResult()
	testResults := p.Parse(path)

	testResultsName, err := cmd.Flags().GetString("name")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return result, err
	}

	if testResultsName != "" {
		logger.Debug("Overriding test results name to %s", testResultsName)
		testResults.Name = testResultsName
		testResults.RegenerateID()
	}

	if testResults.Name == "" {
		logger.Debug("Attempting to name test results automatically")
		testResults.Name = cases.Title(language.English, cases.NoLower).String(fmt.Sprintf("%s Suite", p.GetName()))
	}

	suitePrefix, err := cmd.Flags().GetString("suite-prefix")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return result, err
	}

	if suitePrefix != "" {
		logger.Debug("Prefixing each suite with %s", suitePrefix)
		for suiteIdx := range testResults.Suites {
			testResults.Suites[suiteIdx].Name = fmt.Sprintf("%s %s", suitePrefix, testResults.Suites[suiteIdx].Name)
		}
		testResults.RegenerateID()
	}

	result.TestResults = append(result.TestResults, testResults)

	err = DecorateResults(&result, cmd)
	if err != nil {
		logger.Error("Decorating results failed with error: %v", err)
		return result, err
	}

	return result, nil
}

func DecorateResults(result *parser.Result, cmd *cobra.Command) error {
	omitStdoutForPassed, err := cmd.Flags().GetBool("omit-output-for-passed")
	if err != nil {
		logger.Error("Reading flag omit-output-for-passed failed with error: %v", err)
		return err
	}

	if omitStdoutForPassed {
		for idx := range result.TestResults {
			for suiteIdx := range result.TestResults[idx].Suites {
				for caseIdx := range result.TestResults[idx].Suites[suiteIdx].Tests {
					if result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].State == "passed" {
						result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemErr = ""
						result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemOut = ""
					}
				}
			}
		}
	}

	trimStdoutTo, err := cmd.Flags().GetInt32("trim-output-to")
	if err != nil {
		logger.Error("Reading flag trim-output-to failed with error: %v", err)
		return err
	}

	if trimStdoutTo > 0 {
		for idx := range result.TestResults {
			for suiteIdx := range result.TestResults[idx].Suites {
				if len(result.TestResults[idx].Suites[suiteIdx].SystemErr) > int(trimStdoutTo) {
					result.TestResults[idx].Suites[suiteIdx].SystemErr = result.TestResults[idx].Suites[suiteIdx].SystemErr[:trimStdoutTo]
				}

				if len(result.TestResults[idx].Suites[suiteIdx].SystemOut) > int(trimStdoutTo) {
					result.TestResults[idx].Suites[suiteIdx].SystemOut = result.TestResults[idx].Suites[suiteIdx].SystemOut[:trimStdoutTo]
				}

				for caseIdx := range result.TestResults[idx].Suites[suiteIdx].Tests {
					if len(result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemErr) > int(trimStdoutTo) {
						result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemErr = result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemErr[:trimStdoutTo]
					}

					if len(result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemOut) > int(trimStdoutTo) {
						result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemOut = result.TestResults[idx].Suites[suiteIdx].Tests[caseIdx].SystemOut[:trimStdoutTo]
					}
				}
			}
		}
	}

	return nil
}

// Marshal provides json output for given test results
func Marshal(testResults parser.Result) ([]byte, error) {
	jsonData, err := json.Marshal(testResults)
	if err != nil {
		logger.Error("Marshaling results failed with: %v", err)
		return nil, err
	}
	return jsonData, nil
}

func WriteToFile(data []byte, file *os.File, compress bool) (string, error) {
	return writeToFile(data, file, compress)
}

// WriteToFilePath saves data to given file
func WriteToFilePath(data []byte, path string, compress bool) (string, error) {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		logger.Error("Opening file %s: %v", path, err)
		return "", err
	}
	defer file.Close()

	return writeToFile(data, file, compress)
}

// WriteToTmpFile saves data to temporary file
func WriteToTmpFile(data []byte, compress bool) (string, error) {
	file, err := os.CreateTemp("", "test-results")
	if err != nil {
		logger.Error("Opening file %s: %v", file.Name(), err)
		return "", err
	}
	defer file.Close()

	return writeToFile(data, file, compress)
}

func writeToFile(data []byte, file *os.File, compress bool) (string, error) {
	logger.Info("Saving results to %s", file.Name())

	dataToWrite := data
	if compress {
		compressedData, err := GzipCompress(data)
		if err != nil {
			logger.Error("Output file write failed: %v", err)
			return "", err
		}
		dataToWrite = compressedData
	}

	_, err := file.Write(dataToWrite)
	if err != nil {
		logger.Error("Output file write failed: %v", err)
		return "", err
	}

	if err = file.Sync(); err != nil {
		logger.Error("Output file write failed: %v", err)
		return "", err
	}

	return file.Name(), nil
}

// PushArtifacts publishes artifacts to semaphore artifact storage
func PushArtifacts(level string, file string, destination string, cmd *cobra.Command) (string, error) {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", err
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", err
	}

	artifactsPush := exec.Command("artifact")
	artifactsPush.Args = append(artifactsPush.Args, "push", level, file, "-d", destination)
	if verbose {
		artifactsPush.Args = append(artifactsPush.Args, "-v")
	}

	if force {
		artifactsPush.Args = append(artifactsPush.Args, "-f")
	}

	output, err := artifactsPush.CombinedOutput()

	logger.Info("Pushing artifacts:\n$ %s", artifactsPush.String())

	if err != nil {
		logger.Error("Pushing artifacts failed: %v\n%s", err, string(output))
		return "", err
	}
	return destination, nil
}

// PullArtifacts fetches artifacts from semaphore artifact storage
func PullArtifacts(level string, remotePath string, localPath string, cmd *cobra.Command) (string, error) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", err
	}

	artifactsPush := exec.Command("artifact")
	artifactsPush.Args = append(artifactsPush.Args, "pull", level, remotePath, "-d", localPath)
	if verbose {
		artifactsPush.Args = append(artifactsPush.Args, "-v")
	}

	output, err := artifactsPush.CombinedOutput()

	logger.Info("Pulling artifacts:\n$ %s", artifactsPush.String())

	if err != nil {
		logger.Error("Pulling artifacts failed: %v\n%s", err, string(output))
		return "", err
	}

	return localPath, nil
}

// SetLogLevel sets log level according to flags
func SetLogLevel(cmd *cobra.Command) error {
	trace, err := cmd.Flags().GetBool("trace")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return err
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return err
	}

	if trace {
		logger.SetLevel(logger.TraceLevel)
	} else if verbose {
		logger.SetLevel(logger.DebugLevel)
	}
	return nil
}

// MergeFiles merges all json files found in path into one big blob
func MergeFiles(path string, cmd *cobra.Command) (*parser.Result, error) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}

	_, err = CheckFile(path)
	if err != nil {
		logger.Error(err.Error())
	}

	result := parser.NewResult()
	fun := func(p string, d fs.DirEntry, err error) error {
		if verbose {
			logger.Info("[verbose] Checking file: %s", p)
		}

		if err != nil {
			logger.Info(err.Error())
			return err
		}

		if d.Type().IsDir() {
			return nil
		}

		fs, err := d.Info()
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		if filepath.Ext(fs.Name()) != ".json" {
			return nil
		}

		inFile, err := CheckFile(p)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		newResult, err := Load(inFile)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		if verbose {
			logger.Info("[verbose] File loaded: %s", p)
		}

		result.Combine(*newResult)
		return nil
	}

	err = filepath.WalkDir(path, fun)
	if err != nil {
		logger.Error("Test results dir listing failed: %v", err)
		return nil, err
	}

	return &result, nil
}

// Load ...
func Load(path string) (*parser.Result, error) {
	var result parser.Result
	jsonFile, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	decompressedBytes, err := GzipDecompress(bytes)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decompressedBytes, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func IsGzipCompressed(bytes []byte) bool {
	if len(bytes) < 2 {
		return false
	}

	return bytes[0] == 0x1f && bytes[1] == 0x8b
}

// takes a slice of data bytes, compresses it and replaces with compressed bytes
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return data, err
	}

	err = writer.Close()
	if err != nil {
		return data, err
	}

	return buf.Bytes(), nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	if !IsGzipCompressed(data) {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		logger.Error("Decompression failed: %v", err)
		return data, err
	}
	defer reader.Close()

	newData, err := io.ReadAll(reader)
	if err != nil {
		logger.Error("Decompression failed: %v", err)
		return data, err
	}

	return newData, nil
}
