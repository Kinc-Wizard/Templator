package internals

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RunSupernovaAES runs Supernova with AES encryption
func RunSupernovaAES(inputPath, lang string) (key, iv, encryptedShellcode string, err error) {
	SendDebugMessage("🔐 Starting AES encryption with Supernova...")
	SendDebugMessage(fmt.Sprintf("📁 Input file: %s", inputPath))
	SendDebugMessage(fmt.Sprintf("🌐 Language: %s", lang))
	
	encOut := "/tmp/enc-shellcode.bin"
	SendDebugMessage(fmt.Sprintf("📄 Output file: %s", encOut))
	
	cmd := exec.Command(
		"tools/native/Supernova/Supernova",
		"-enc", "AES",
		"-input", inputPath,
		"-key", "32",
		"-lang", lang,
	)
	
	SendDebugMessage("⚙️ Running Supernova encryption tool...")
	
	outfile, err := os.Create(encOut)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to create output file: %v", err))
		return "", "", "", fmt.Errorf("failed to create enc-shellcode.bin: %v", err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Supernova encryption failed: %v", err))
		return "", "", "", fmt.Errorf("Supernova error: %v", err)
	}

	SendDebugMessage("✅ Supernova encryption completed successfully")
	SendDebugMessage("🔍 Parsing encryption output...")

	// Now parse enc-shellcode.bin for key, iv, and encrypted payload
	key, iv, encryptedShellcode, err = parseSupernovaOutput(encOut)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to parse encryption output: %v", err))
	} else {
		SendDebugMessage("✅ Encryption output parsed successfully")
		SendDebugMessage(fmt.Sprintf("🔑 Key length: %d bytes", len(key)))
		SendDebugMessage(fmt.Sprintf("🔐 IV length: %d bytes", len(iv)))
		SendDebugMessage(fmt.Sprintf("🔒 Encrypted shellcode length: %d bytes", len(encryptedShellcode)))
	}
	return
}

// RunSupernovaEncryption runs Supernova with dynamic encryption algorithm
func RunSupernovaEncryption(inputPath, lang, encAlgo string) (key, iv, encryptedShellcode string, err error) {
	SendDebugMessage(fmt.Sprintf("🔐 Starting %s encryption with Supernova...", encAlgo))
	SendDebugMessage(fmt.Sprintf("📁 Input file: %s", inputPath))
	SendDebugMessage(fmt.Sprintf("🌐 Language: %s", lang))
	SendDebugMessage(fmt.Sprintf("🔒 Algorithm: %s", encAlgo))
	
	encOut := "/tmp/enc-shellcode.bin"
	SendDebugMessage(fmt.Sprintf("📄 Output file: %s", encOut))
	
	cmd := exec.Command(
		"tools/native/Supernova/Supernova",
		"-enc", encAlgo,
		"-input", inputPath,
		"-key", "32",
		"-lang", lang,
	)
	
	SendDebugMessage("⚙️ Running Supernova encryption tool...")
	
	outfile, err := os.Create(encOut)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to create output file: %v", err))
		return "", "", "", fmt.Errorf("failed to create enc-shellcode.bin: %v", err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Supernova encryption failed: %v", err))
		return "", "", "", fmt.Errorf("Supernova error: %v", err)
	}

	SendDebugMessage("✅ Supernova encryption completed successfully")
	SendDebugMessage("🔍 Parsing encryption output...")

	key, iv, encryptedShellcode, err = parseSupernovaOutput(encOut)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to parse encryption output: %v", err))
	} else {
		SendDebugMessage("✅ Encryption output parsed successfully")
		SendDebugMessage(fmt.Sprintf("🔑 Key length: %d bytes", len(key)))
		SendDebugMessage(fmt.Sprintf("🔐 IV length: %d bytes", len(iv)))
		SendDebugMessage(fmt.Sprintf("🔒 Encrypted shellcode length: %d bytes", len(encryptedShellcode)))
	}
	return
}

// parseSupernovaOutput parses Supernova output for key, iv, and encrypted shellcode
func parseSupernovaOutput(path string) (key, iv, encryptedShellcode string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read enc-shellcode.bin: %v", err)
	}
	text := string(data)

	// Parse key
	keyStart := strings.Index(text, "[+] Generated key")
	if keyStart == -1 {
		return "", "", "", fmt.Errorf("key not found in Supernova output")
	}
	keyLine := text[keyStart : strings.Index(text[keyStart:], "\n")+keyStart]
	keyBytes := extractHexBytes(keyLine)
	key = formatCArray("key", 32, keyBytes)

	// Parse IV
	ivStart := strings.Index(text, "[+] Generated IV")
	if ivStart == -1 {
		return "", "", "", fmt.Errorf("IV not found in Supernova output")
	}
	ivLine := text[ivStart : strings.Index(text[ivStart:], "\n")+ivStart]
	ivBytes := extractHexBytes(ivLine)
	iv = formatCArray("iv", 16, ivBytes)

	// Parse encrypted shellcode block (unsigned char shellcode[] = { ... };)
	encStart := strings.Index(text, "unsigned char shellcode[]")
	if encStart == -1 {
		return "", "", "", fmt.Errorf("encrypted shellcode not found")
	}
	encEnd := strings.Index(text[encStart:], "};")
	if encEnd == -1 {
		return "", "", "", fmt.Errorf("encrypted shellcode block not closed")
	}
	encEnd += encStart + 2 // include "};"
	encryptedShellcode = text[encStart:encEnd]

	return key, iv, encryptedShellcode, nil
}

// extractHexBytes extracts hex bytes from a line like: byte(0x09) => 9, ...
func extractHexBytes(line string) []byte {
	var bytesArr []byte
	parts := strings.Split(line, "byte(0x")
	for _, part := range parts[1:] {
		hexEnd := strings.Index(part, ")")
		if hexEnd == -1 {
			continue
		}
		hexStr := part[:hexEnd]
		val, err := strconv.ParseUint(hexStr, 16, 8)
		if err == nil {
			bytesArr = append(bytesArr, byte(val))
		}
	}
	return bytesArr
}

// formatCArray formats bytes as C array
func formatCArray(name string, size int, arr []byte) string {
	var b strings.Builder
	fmt.Fprintf(&b, "unsigned char %s[%d] = {\n", name, size)
	for i, v := range arr {
		if i%8 == 0 {
			b.WriteString("    ")
		}
		fmt.Fprintf(&b, "0x%02x", v)
		if i < len(arr)-1 {
			b.WriteString(", ")
		}
		if (i+1)%8 == 0 {
			b.WriteString("\n")
		}
	}
	if len(arr)%8 != 0 {
		b.WriteString("\n")
	}
	b.WriteString("};\n")
	return b.String()
}
