package parsers

import (
	"testing"
	
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// TestIDGenerationChain verifies that IDs are generated deterministically
// and properly chain from parent to child
func TestIDGenerationChain(t *testing.T) {
	t.Run("Deterministic ID Generation", func(t *testing.T) {
		// Create two identical test results
		tr1 := parser.NewTestResults()
		tr1.Name = "Test Suite"
		tr1.Framework = "golang"
		tr1.EnsureID()
		
		tr2 := parser.NewTestResults()
		tr2.Name = "Test Suite"
		tr2.Framework = "golang"
		tr2.EnsureID()
		
		if tr1.ID != tr2.ID {
			t.Errorf("Same input should produce same ID: %s != %s", tr1.ID, tr2.ID)
		}
	})
	
	t.Run("Parent-Child ID Chaining", func(t *testing.T) {
		// Create parent
		tr := parser.NewTestResults()
		tr.Name = "Parent"
		tr.Framework = "test"
		tr.EnsureID()
		parentID := tr.ID
		
		// Create child suite
		suite := parser.NewSuite()
		suite.Name = "Child Suite"
		suite.EnsureID(tr)
		suiteID := suite.ID
		
		// Create grandchild test
		test := parser.NewTest()
		test.Name = "Test Case"
		test.EnsureID(suite)
		testID := test.ID
		
		// Verify all IDs are different
		if parentID == suiteID || suiteID == testID || parentID == testID {
			t.Error("Parent and child IDs should be different")
		}
		
		// Change parent and verify children change
		tr.Name = "Different Parent"
		tr.RegenerateID()
		
		suite2 := parser.NewSuite()
		suite2.Name = "Child Suite" // Same name as before
		suite2.EnsureID(tr)
		
		if suite.ID == suite2.ID {
			t.Error("Child ID should change when parent ID changes")
		}
	})
	
	t.Run("Sibling IDs Are Different", func(t *testing.T) {
		tr := parser.NewTestResults()
		tr.Name = "Parent"
		tr.EnsureID()
		
		// Create two sibling suites with different names
		suite1 := parser.NewSuite()
		suite1.Name = "Suite A"
		suite1.EnsureID(tr)
		
		suite2 := parser.NewSuite()
		suite2.Name = "Suite B"
		suite2.EnsureID(tr)
		
		if suite1.ID == suite2.ID {
			t.Error("Sibling suites should have different IDs")
		}
		
		// Create two tests in the same suite
		test1 := parser.NewTest()
		test1.Name = "Test 1"
		test1.EnsureID(suite1)
		
		test2 := parser.NewTest()
		test2.Name = "Test 2"
		test2.EnsureID(suite1)
		
		if test1.ID == test2.ID {
			t.Error("Sibling tests should have different IDs")
		}
	})
	
	t.Run("ID Determinism Across Runs", func(t *testing.T) {
		// This test verifies that the same structure always produces the same IDs
		generateStructure := func() (string, string, string) {
			tr := parser.NewTestResults()
			tr.Name = "Consistent"
			tr.Framework = "framework"
			tr.EnsureID()
			
			suite := parser.NewSuite()
			suite.Name = "MySuite"
			suite.EnsureID(tr)
			
			test := parser.NewTest()
			test.Name = "MyTest"
			test.Classname = "MyClass"
			test.EnsureID(suite)
			
			return tr.ID, suite.ID, test.ID
		}
		
		// Generate IDs multiple times
		trID1, suiteID1, testID1 := generateStructure()
		trID2, suiteID2, testID2 := generateStructure()
		trID3, suiteID3, testID3 := generateStructure()
		
		// All should be identical
		if trID1 != trID2 || trID2 != trID3 {
			t.Errorf("TestResults IDs not consistent: %s, %s, %s", trID1, trID2, trID3)
		}
		if suiteID1 != suiteID2 || suiteID2 != suiteID3 {
			t.Errorf("Suite IDs not consistent: %s, %s, %s", suiteID1, suiteID2, suiteID3)
		}
		if testID1 != testID2 || testID2 != testID3 {
			t.Errorf("Test IDs not consistent: %s, %s, %s", testID1, testID2, testID3)
		}
	})
}