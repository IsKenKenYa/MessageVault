package contracttest

import "testing"

func TestSQLiteContract(t *testing.T) {
	RunContractTests(t, NewSQLiteTestProvider)
}
