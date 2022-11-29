package mclifilesystem

// https://www.devdungeon.com/content/working-files-go

type SetFileContentType func(string, string) (int, error)

var SetFile SetFileContentType = func(string, string) (int, error) {
	return 0, nil
}

func (sc SetFileContentType) SetContent(filePath string, content string) (int, error) {
	f, closer, err := getFileForRW(filePath)
	if err != nil {
		return 0, err
	}
	defer closer()

	n, err := f.Write([]byte(content))
	if err != nil {
		return 0, err
	}
	return n, nil
}
