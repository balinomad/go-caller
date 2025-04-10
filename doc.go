/*
Package caller provides utilities to extract source code location
information (file, line, function, and package) for the current
or specified call frame.
It is designed for use in logging, error reporting, and debugging
with a lightweight and idiomatic API. Caller captures runtime metadata
using the Go runtime and formats it in a developer-friendly way.

Example usage:

	import "github.com/balinomad/go-caller"

	func someFunc() {
		c := caller.Immediate()
		fmt.Println("Caller location:", c.Location())
		fmt.Println("Short:", c.ShortLocation())
		fmt.Println("Function:", c.Function())
		fmt.Println("Package:", c.PackageName())
		data, err := json.Marshal(c)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("JSON:", string(data))
	}
*/
package caller
