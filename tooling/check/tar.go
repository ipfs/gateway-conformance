package check

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"strings"
)

var _ Check[[]byte] = &CheckIsTarFile{}

type CheckIsTarFile struct {
	fileNames        []string
	filesWithContent map[string]string
}

func IsTarFile() *CheckIsTarFile {
	return &CheckIsTarFile{
		fileNames:        []string{},
		filesWithContent: map[string]string{},
	}
}

func (c *CheckIsTarFile) HasFile(format string, a ...interface{}) *CheckIsTarFile {
	fileName := fmt.Sprintf(format, a...)
	c.fileNames = append(c.fileNames, fileName)
	return c
}

func (c *CheckIsTarFile) HasFileWithContent(fileName, content string) *CheckIsTarFile {
	c.filesWithContent[fileName] = content
	return c
}

func (c *CheckIsTarFile) Check(v []byte) CheckOutput {
	r := bytes.NewReader(v)
	tr := tar.NewReader(r)

	foundFiles := make(map[string]bool)
	foundFilesWithContent := make(map[string]bool)
	fileContents := make(map[string]string)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("failed to read tar header: %v", err),
			}
		}

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, tr)
		if err != nil {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("failed to read file '%s' content: %v", hdr.Name, err),
			}
		}
		fileContents[hdr.Name] = buf.String()

		for _, fileName := range c.fileNames {
			if hdr.Name == fileName {
				foundFiles[fileName] = true
			}
		}

		if content, ok := c.filesWithContent[hdr.Name]; ok {
			if buf.String() == content {
				foundFilesWithContent[hdr.Name] = true
			}
		}
	}

	for _, fileName := range c.fileNames {
		if !foundFiles[fileName] {
			var fileList strings.Builder
			for name := range fileContents {
				fileList.WriteString(fmt.Sprintf("'%s', ", name))
			}
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("file '%s' not found in tar archive. Found files: [%s]", fileName, fileList.String()),
			}
		}
	}

	for fileName, content := range c.filesWithContent {
		if !foundFilesWithContent[fileName] {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("file '%s' with expected content '%s' not found in tar archive. Actual content: '%s'", fileName, content, fileContents[fileName]),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}
