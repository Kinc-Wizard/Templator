package internals

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRandomFilename generates a random filename
func GenerateRandomFilename() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// TryAutoParseShellcode attempts to automatically parse shellcode in various formats
func TryAutoParseShellcode(path string, language string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	text := string(data)

	// Detect existing C# format
	if strings.Contains(text, "byte[] buf = new byte[") {
		// The shellcode is already in C# format, return as is
		return text, nil
	}

	// If the content starts with "byte[] buf = new byte[]" without specified size
	if strings.Contains(text, "byte[] buf = new byte[] {") {
		return text, nil
	}

	// The rest of the code for other formats...
	if strings.Contains(text, "unsigned char") && strings.Contains(text, "buf[]") {
		// Already formatted C
		if language == "csharp" {
			// Convert C format to C#
			return convertCToCSharpArray(text)
		}
		return text, nil
	}

	if strings.Contains(text, "\\x") && !strings.Contains(text, "{") && !strings.Contains(text, "0x") {
		if language == "csharp" {
			// Format for C#
			bytes := strings.Split(text, "\\x")
			csharpArray := "byte[] buf = new byte[] {"
			for i, b := range bytes[1:] { // Skip first empty element
				if i%12 == 0 {
					csharpArray += "\n\t"
				}
				csharpArray += "0x" + b
				if i < len(bytes)-2 {
					csharpArray += ", "
				}
			}
			csharpArray += "\n};"
			return csharpArray, nil
		} else {
			// Format for C
			buf := fmt.Sprintf("unsigned char buf[] = \n%s;\n", text)
			buf += "unsigned int buf_len = sizeof(buf) - 1;\n"
			return buf, nil
		}
	}

	return shellcodeToCArray(path, language)
}

// shellcodeToCArray converts shellcode file to C array format
func shellcodeToCArray(path string, language string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if language == "csharp" {
		csharpArray := "byte[] buf = new byte[] {"
		for i, b := range data {
			if i%12 == 0 {
				csharpArray += "\n\t"
			}
			csharpArray += fmt.Sprintf("0x%02x", b)
			if i < len(data)-1 {
				csharpArray += ", "
			}
		}
		csharpArray += "\n};"
		return csharpArray, nil
	}

	if language == "rust" {
		rustArray := "let buf: [u8; " + fmt.Sprintf("%d", len(data)) + "] = ["
		for i, b := range data {
			if i%16 == 0 {
				rustArray += "\n\t"
			}
			rustArray += fmt.Sprintf("0x%02x", b)
			if i < len(data)-1 {
				rustArray += ", "
			}
		}
		rustArray += "\n];"
		return rustArray, nil
	}

	// Default C format
	cArray := "unsigned char buf[] = {"
	for i, b := range data {
		if i%12 == 0 {
			cArray += "\n\t"
		}
		cArray += fmt.Sprintf("0x%02x", b)
		if i < len(data)-1 {
			cArray += ", "
		}
	}
	cArray += "\n};"
	return cArray, nil
}

// convertCToCSharpArray converts C array format to C# array format
func convertCToCSharpArray(cArray string) (string, error) {
	// Extract bytes between braces
	start := strings.Index(cArray, "{")
	end := strings.LastIndex(cArray, "}")
	if start == -1 || end == -1 {
		return "", fmt.Errorf("invalid C array format")
	}

	bytesStr := cArray[start+1 : end]
	bytesStr = strings.ReplaceAll(bytesStr, "\n", "")
	bytesStr = strings.ReplaceAll(bytesStr, "\t", "")
	bytesStr = strings.ReplaceAll(bytesStr, " ", "")
	byteVals := strings.Split(bytesStr, ",")
	csharpArray := "byte[] buf = new byte[] {"
	for i, val := range byteVals {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		if i%12 == 0 {
			csharpArray += "\n\t"
		}
		csharpArray += val
		if i < len(byteVals)-1 {
			csharpArray += ", "
		}
	}
	csharpArray += "\n};"
	return csharpArray, nil
}

// SaveUploadedFile saves an uploaded file to the uploads directory
func SaveUploadedFile(file io.Reader, filename string) (string, error) {
	tmpShellcodePath := filepath.Join("uploads", filename)
	outFile, err := os.Create(tmpShellcodePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}
	defer outFile.Close()
	
	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", fmt.Errorf("error copying file: %v", err)
	}
	
	return tmpShellcodePath, nil
}
