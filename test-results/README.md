# Test results [![SemaphoreCI](https://semaphore.semaphoreci.com/badges/test-results-cli.svg)](https://semaphore.semaphoreci.com/projects/test-results-cli)

Semaphore collects XML test reports and uses them to provide insight into your pipelines.

With test reports, you enable your team to get an effective and consistent view of your CI/CD test suite across different test frameworks and stages in a CI/CD workflow. You get a clear failure report for each executed pipeline. Failures are extracted and highlighted, while the rest of the suite is available for analysis.

The test-results command-line interface (CLI) is a tool that helps you compile and process [JUnit XML](https://github.com/windyroad/JUnit-Schema/blob/master/JUnit.xsd) files. The output of the test results CLI is a report in JSON format.

This CLI is distributed as a part of the [Semaphore toolbox](https://github.com/semaphoreci/toolbox), and it is available in all Semaphore jobs.

The main purpose of the CLI is to:

- compile and publish JUnit XML files into a JSON report
- merge multiple JSON reports into a single summary report

> [!NOTE]
>
> Starting from version `0.7.0`, test reports are compressed using gzip to optimize storage and handling. To maintain backward compatibility with existing systems and processes, the compressed files will continue to use the `.json` file extension. This approach ensures seamless integration with tools and workflows that expect files in this format.
>
> However, please be aware that while these files retain the `.json` extension, they are in a compressed format and will need to be decompressed using gzip-compatible tools before they can be read as standard JSON.
>
> For users who prefer to work with uncompressed reports or for systems that require non-compressed files, we've introduced a `--no-compress` option. This can be used to generate and upload test reports in the traditional, uncompressed JSON format.

## Compiling and publishing JUnit XML files

Given your JUnit XML report is named `results.xml` you can run the following command to generate a report:

```bash
test-results publish results.xml
```

The above command parses the content of the `results.xml` file, and publishes the results to Semaphore.

While parsing the content, the CLI tries to find the best parser for your result type. The following test runners have a dedicated parser:

- exunit
- golang
- mocha
- rspec
- phpunit

If a dedicated parser is not found, the CLI will parse the file using a generic parser. The generic parser uses [JUnit XML Schema](https://github.com/windyroad/JUnit-Schema/blob/master/JUnit.xsd) definition to extract data from the report.

The parser can be selected manually by using the `--parser` option.

```bash
test-results publish --parser exunit results.xml
```

The name of the generated report is based on the selected parser. If you want to overwrite this you can use the `--name` option:

```bash
test-results publish --name "Unit Tests" results.xml
```

The generated tests in a report will sometimes contain a prefix in the name. For example `Elixir.MyTest`. If you want to remove the `Elixir` prefix from the test names you can use `--suite-prefix` option:

```bash
test-results publish --suite-prefix "Elixir." results.xml
```

## Multiple reports from one job

If your job generates multiple reports: `integration.xml`, `unit.xml` you can use this command to merge and publish them

```bash
test-results publish integration.xml unit.xml
```

In addition, each report is published separately to artifact storage as a `junit-<index>.xml`. `<index>` is a number starting from 0 that corresponds to the order of the report passed to the command line.

## Merging multiple JSON reports into a single summary report

If you have multiple jobs in your pipeline that generate test results, you can merge them into a single report with the following command

```bash
test-results gen-pipeline-report
```

The above command assumes you are running it in a semaphore pipeline. As it uses `SEMAPHORE_PIPELINE_ID` environment variable to identify the pipeline and fetch the job level reports.

## Where are test reports stored?

The test results CLI uses the [Semaphore Artifact Storage](https://docs.semaphoreci.com/essentials/artifacts/) to store the test reports:

- the `test-results publish` command stores the report in the `test-results/junit.json` file on a job level
- the `test-results gen-pipeline-report` command stores the report in the `test-results/${SEMAPHORE_PIPELINE_ID}.json` file on a workflow level

Expiration date of the artifacts can be controlled via [Retention Policies](https://docs.semaphoreci.com/essentials/artifacts/#artifact-retention-policies).

## Skip uploading raw JUnit XML files

By default, `test-results publish` will upload the raw JUnit XML file alongside the JSON report to the artifact storage. This can be disabled with the `--no-raw` option:

```bash
test-results publish --no-raw results.xml
```

## Overwrite existing reports

By default `test-results publish` and `test-results gen-pipeline-report` will fail if the report is already present in the artifact storage. This behaviour can be disabled with the `--force` option:

```bash
# Publish the report
test-results publish results.xml

#...

# other-results will overwrite results
test-results publish --force other-results.xml
```

## Using the CLI on a local machine

Latest CLI binaries are available [here](https://github.com/semaphoreci/test-results/releases/latest).
