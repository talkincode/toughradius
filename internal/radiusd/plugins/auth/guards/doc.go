// Package guards provides pipeline guards that shape authentication response
// behavior.
//
// Guard hooks run around Access-Reject handling to apply cross-cutting
// controls, such as deterministic reject delays, without duplicating that logic
// in every validator or checker.
package guards
