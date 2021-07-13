package utils

import "runtime"

// ErrSliceSortAlphabetical is a helper type that can be used with sort.Sort to sort a slice of errors in alphabetical
// order. Usage is simple just do sort.Sort(ErrSliceSortAlphabetical([]error{})).
type ErrSliceSortAlphabetical []error

func (s ErrSliceSortAlphabetical) Len() int { return len(s) }

func (s ErrSliceSortAlphabetical) Less(i, j int) bool { return s[i].Error() < s[j].Error() }

func (s ErrSliceSortAlphabetical) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetExpectedErrTxt returns error text for expected errs.
func GetExpectedErrTxt(err string) string {
	switch err {
	case "pathnotfound":
		switch runtime.GOOS {
		case windows:
			return "open %s: The system cannot find the path specified."
		default:
			return errFmtLinuxNotFound
		}
	case "filenotfound":
		switch runtime.GOOS {
		case windows:
			return "open %s: The system cannot find the file specified."
		default:
			return errFmtLinuxNotFound
		}
	case "yamlisdir":
		switch runtime.GOOS {
		case windows:
			return "read %s: The handle is invalid."
		default:
			return "read %s: is a directory"
		}
	}

	return ""
}
