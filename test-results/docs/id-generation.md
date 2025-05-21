## ID generation

This PR introduces `id` generator for tests results, test suites, and tests.

`id`'s are being generated in form of UUID strings.

To generate consistent `id`'s between builds following method is implemented for all parsers:

- ID generation for `Test results`(top-level element)

  1. If the element has an ID, generate UUID based on that ID
  2. If the element doesn't have an ID - generate UUID based on the `name` attribute
  3. If the element has a framework name - generate UUID based on the `name` attribute and `framework`
  4. Otherwise, generate uuid based on an empty string `""`

- ID generation for `Suites`

  The same rules apply as for `Test results` however every `Suite ID` is namespaced by `Test results` ID

- ID generation for `Tests`

  The same rules apply as for `Test results` however every `Test ID` is namespaced by `Suite` ID and `Test classname` if present.
  If a test is failed/errored the state is also added as namespace, as failed/errored cases can happen simultaneously in the same suite.

For generating IDs we're using [UUID v3 generator](https://pkg.go.dev/github.com/google/uuid#NewMD5).