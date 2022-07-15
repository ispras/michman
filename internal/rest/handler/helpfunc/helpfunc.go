package helpfunc

// DeleteSpaces deletes all spaces in given string
func DeleteSpaces(valStr string) string {
	resStr := ""
	for _, ch := range valStr {
		if ch != ' ' {
			resStr += string(ch)
		}
	}
	return resStr
}

// MakeLogFilePath makes unified paths for custom and regular paths for the location of directories of files with logs
func MakeLogFilePath(filename string, LogsFilePath string) string {
	if LogsFilePath[0] == '/' {
		return LogsFilePath + "/" + filename
	}
	return "./" + LogsFilePath + "/" + filename
}
