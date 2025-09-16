package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/compression"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/semaphoreci/toolbox/test-results/pkg/parsers"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ArtifactStats struct {
	Operations int
	FileCount  int
	TotalSize  int64
}

// FileParserPair represents a file and its optional parser
type FileParserPair struct {
	Path   string
	Parser string // empty means use default/auto-detect
}

// ParseFileArgs parses arguments like "file.xml:parser" into pairs
func ParseFileArgs(args []string) []FileParserPair {
	var pairs []FileParserPair
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 2 {
			// Has explicit parser
			pairs = append(pairs, FileParserPair{Path: parts[0], Parser: parts[1]})
		} else {
			// Use default parser logic
			pairs = append(pairs, FileParserPair{Path: parts[0], Parser: ""})
		}
	}
	return pairs
}

// FindParserForFile finds parser with optional override
func FindParserForFile(pair FileParserPair, cmd *cobra.Command) (parser.Parser, error) {
	if pair.Parser != "" {
		// Use specified parser
		p, err := parsers.FindParser(pair.Parser, pair.Path)
		if err != nil {
			logger.Error("Could not find specified parser %s: %v", pair.Parser, err)
			return nil, fmt.Errorf("parser %s not found", pair.Parser)
		}
		logger.Info("Using %s parser (explicitly specified)", p.GetName())
		return p, nil
	}
	// Use default logic (from -p flag or auto-detect)
	return FindParser(pair.Path, cmd)
}

// EmptyResult returns an empty result structure
func EmptyResult() *parser.Result {
	return &parser.Result{
		TestResults: []parser.TestResults{},
	}
}

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
		compressedData, err := compression.GzipCompress(data)
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
func PushArtifacts(level string, file string, destination string, cmd *cobra.Command) (string, *ArtifactStats, error) {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", nil, err
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", nil, err
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
		return "", nil, err
	}

	stats := parseArtifactStats(string(output))

	return destination, stats, nil
}

// PullArtifacts fetches artifacts from semaphore artifact storage
func PullArtifacts(level string, remotePath string, localPath string, cmd *cobra.Command) (string, *ArtifactStats, error) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error("Reading flag error: %v", err)
		return "", nil, err
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
		return "", nil, err
	}

	stats := parseArtifactStats(string(output))

	return localPath, stats, nil
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

	// Get a reader that handles both compressed and uncompressed files
	reader, closeFunc, err := compression.GzipDecompress(jsonFile)
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	// Stream decode JSON directly from reader
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func parseArtifactStats(output string) *ArtifactStats {
	stats := &ArtifactStats{}

	fileCountRegex := regexp.MustCompile(`(?:Pushed|Pulled)\s+(\d+)\s+files?`)
	sizeRegex := regexp.MustCompile(`Total\s+of\s+([\d.]+)\s*([KMGT]?B)`)

	if matches := fileCountRegex.FindStringSubmatch(output); len(matches) > 1 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			stats.FileCount = count
		}
	}

	if matches := sizeRegex.FindStringSubmatch(output); len(matches) > 1 {
		if size, err := parseSize(matches[1], matches[2]); err == nil {
			stats.TotalSize = size
		}
	}

	return stats
}

func parseSize(sizeStr string, unit string) (int64, error) {
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}

	multiplier := int64(1)
	switch strings.ToUpper(unit) {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(size * float64(multiplier)), nil
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func Pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func DisplayTransferSummary(pullStats *ArtifactStats, pushStats *ArtifactStats) {
	totalOps := pullStats.Operations + pushStats.Operations
	if totalOps > 0 {
		logger.Info("[test-results] Artifact transfers:")

		if pullStats.Operations > 0 {
			if pullStats.FileCount > 0 || pullStats.TotalSize > 0 {
				logger.Info("  → Pulled: %d %s, %d %s, %s",
					pullStats.Operations, Pluralize(pullStats.Operations, "operation", "operations"),
					pullStats.FileCount, Pluralize(pullStats.FileCount, "file", "files"),
					FormatBytes(pullStats.TotalSize))
			} else {
				logger.Info("  → Pulled: %d %s",
					pullStats.Operations, Pluralize(pullStats.Operations, "operation", "operations"))
			}
		}

		if pushStats.Operations > 0 {
			if pushStats.FileCount > 0 || pushStats.TotalSize > 0 {
				logger.Info("  ← Pushed: %d %s, %d %s, %s",
					pushStats.Operations, Pluralize(pushStats.Operations, "operation", "operations"),
					pushStats.FileCount, Pluralize(pushStats.FileCount, "file", "files"),
					FormatBytes(pushStats.TotalSize))
			} else {
				logger.Info("  ← Pushed: %d %s",
					pushStats.Operations, Pluralize(pushStats.Operations, "operation", "operations"))
			}
		}

		totalFiles := pullStats.FileCount + pushStats.FileCount
		totalSize := pullStats.TotalSize + pushStats.TotalSize
		if totalFiles > 0 || totalSize > 0 {
			logger.Info("  = Total: %d %s, %d %s, %s",
				totalOps, Pluralize(totalOps, "operation", "operations"),
				totalFiles, Pluralize(totalFiles, "file", "files"),
				FormatBytes(totalSize))
		} else {
			logger.Info("  = Total: %d %s",
				totalOps, Pluralize(totalOps, "operation", "operations"))
		}
	}
}
