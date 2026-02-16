package check

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling/tmpl"
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

func (c *CheckIsTarFile) HasFile(format string, a ...any) *CheckIsTarFile {
	fileName := tmpl.Fmt(format, a...)
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

	searchedFiles := make(map[string]bool)
	filenames := make([]string, 0, len(c.fileNames))

	for _, fileName := range c.fileNames {
		searchedFiles[fileName] = false
	}

	for fileName := range c.filesWithContent {
		searchedFiles[fileName] = false
	}

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

		filenames = append(filenames, hdr.Name)

		if _, ok := searchedFiles[hdr.Name]; ok {
			searchedFiles[hdr.Name] = true
		}

		if _, ok := c.filesWithContent[hdr.Name]; ok {
			content := buf.String()

			if content != c.filesWithContent[hdr.Name] {
				return CheckOutput{
					Success: false,
					Reason:  fmt.Sprintf("file '%s' with expected content '%s' not found in tar archive. Actual content: '%s'", hdr.Name, c.filesWithContent[hdr.Name], content),
				}
			}
		}
	}

	for name, found := range searchedFiles {
		if !found {
			allFiles := strings.Join(filenames, ", ")

			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("file '%s' not found in tar archive. Found files: [%s]", name, allFiles),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}
