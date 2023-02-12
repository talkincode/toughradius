package wfs

// CombinedPolicy allows to join multiple policies together
type CombinedPolicy struct {
	Policies []Policy
}

// Comply method returns true only when all joined policies are complied
func (p CombinedPolicy) Comply(path FileID, operation int) bool {
	for _, el := range p.Policies {
		if el.Comply(path, operation) == false {
			return false
		}
	}
	return true
}

// ReadOnlyPolicy allows read access and blocks any modifications
type ReadOnlyPolicy struct{}

// Comply method returns true for read operations
func (p ReadOnlyPolicy) Comply(path FileID, operation int) bool {
	return operation == ReadOperation
}

// AllowPolicy allows all operations
type AllowPolicy struct{}

// Comply method returns true
func (p AllowPolicy) Comply(path FileID, operation int) bool {
	return true
}

// DenyPolicy allows all operations
type DenyPolicy struct{}

// Comply method returns false
func (p DenyPolicy) Comply(path FileID, operation int) bool {
	return false
}
