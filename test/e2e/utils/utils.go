package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/onsi/gomega"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func UnmarshalCallToolResult[T any](output []byte) T {
	var result mcp.CallToolResult
	err := result.UnmarshalJSON(output)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(result.StructuredContent).NotTo(gomega.BeEmpty())

	jsonOutput, err := json.Marshal(result.StructuredContent)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var resultInTFormat T
	err = json.Unmarshal(jsonOutput, &resultInTFormat)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return resultInTFormat
}

func GetTestdataPath(relativePath string) string {
	_, thisFile, _, ok := runtime.Caller(1)
	gomega.Expect(ok).To(gomega.BeTrue())

	path, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), relativePath))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = os.Stat(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return path
}
