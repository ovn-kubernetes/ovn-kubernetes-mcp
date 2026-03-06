package utils

import (
	"encoding/json"
	"path/filepath"
	"runtime"

	. "github.com/onsi/gomega"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func UnmarshalCallToolResult[T any](output []byte) T {
	var result mcp.CallToolResult
	err := result.UnmarshalJSON(output)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.StructuredContent).NotTo(BeEmpty())

	jsonOutput, err := json.Marshal(result.StructuredContent)
	Expect(err).NotTo(HaveOccurred())

	var resultInTFormat T
	err = json.Unmarshal(jsonOutput, &resultInTFormat)
	Expect(err).NotTo(HaveOccurred())

	return resultInTFormat
}

func GetTestdataPath(relativePath string) string {
	_, thisFile, _, ok := runtime.Caller(1)
	if !ok {
		Expect(ok).To(BeTrue())
	}
	path, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), relativePath))
	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	return path
}
