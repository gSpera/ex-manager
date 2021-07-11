// ex is a library to build and manage exploits for A/D CTF.
//
// A/D CTFs requires a moltitude of attacks towards multiple targets
// for stoling flags, ex (and it cmds) provides a set of tools to execute
// those attacks to all targets and submit the flags that are found.
//
// The main component is a Session, a session manages a CTF
// and manages all Services found in a CTF, the Submitter, the Flags, and the Workers,
// a service can have multiple exploits, those are the attacks
// that will be used against the targets
package ex

type (
	Target    = string
	FlagValue = string
)
