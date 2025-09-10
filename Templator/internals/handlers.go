package internals

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Payload represents a generated payload
type Payload struct {
	Filename   string `json:"filename"`
	Template   string `json:"template"`
	Language   string `json:"language,omitempty"`
	Arch       string `json:"arch"`
	Created    string `json:"created"`
	Obfuscated bool   `json:"obfuscated"`
	Encrypted  string `json:"encrypted,omitempty"`
	Process    string `json:"process_name,omitempty"`
}

// WebTemplates holds the parsed HTML templates
var WebTemplates *template.Template

// InitWebTemplates initializes the web templates
func InitWebTemplates() {
	WebTemplates = template.Must(template.ParseGlob("templates_web/*.html"))
}

// UploadHandler handles file uploads and payload generation
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	SendDebugMessage("🚀 Starting payload generation process...")
	
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Form parsing error: %v", err))
		http.Redirect(w, r, "/?message=Form parsing error&status=error", http.StatusSeeOther)
		return
	}

	file, handler, err := r.FormFile("shellcode")
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Shellcode file not found: %v", err))
		http.Redirect(w, r, "/?message=Shellcode file not found&status=error", http.StatusSeeOther)
		return
	}
	defer file.Close()

	arch := r.FormValue("arch")
	templateName := r.FormValue("template")
	selectedLanguage := r.FormValue("language")
	processName := r.FormValue("process_name")
	processPath := r.FormValue("process_path")

	SendDebugMessage(fmt.Sprintf("📋 Generation parameters:"))
	SendDebugMessage(fmt.Sprintf("  🏗️  Architecture: %s", arch))
	SendDebugMessage(fmt.Sprintf("  📄 Template: %s", templateName))
	if processName != "" {
		SendDebugMessage(fmt.Sprintf("  🎯 Process name: %s", processName))
	}
	if processPath != "" {
		SendDebugMessage(fmt.Sprintf("  📂 Process path: %s", processPath))
	}

	// Get template info from availableTemplates
	var selectedTemplate Template
	// Prefer exact match on name + language when provided
	if selectedLanguage != "" {
		for _, t := range AvailableTemplates {
			if t.Name == templateName && t.Language == selectedLanguage {
				selectedTemplate = t
				break
			}
		}
	}
	// Fallback: first match by name
	if selectedTemplate.Path == "" {
		for _, t := range AvailableTemplates {
			if t.Name == templateName {
				selectedTemplate = t
				break
			}
		}
	}

	if selectedTemplate.Path == "" {
		SendDebugMessage(fmt.Sprintf("❌ Template not found: %s", templateName))
		http.Error(w, "Template not found", http.StatusBadRequest)
		return
	}

	SendDebugMessage(fmt.Sprintf("✅ Template found: %s (%s)", selectedTemplate.Name, selectedTemplate.Language))

	// Validation
	if strings.Contains(templateName, "spawn-process") && processName == "" {
		SendDebugMessage("❌ Process name required for spawn-process template")
		http.Redirect(w, r, "/?message=Process name required for this template&status=error", http.StatusSeeOther)
		return
	}

	if selectedTemplate.NeedsProcessPath && processPath == "" {
		SendDebugMessage("❌ Process path required for this template")
		http.Redirect(w, r, "/?message=Process path required for this template&status=error", http.StatusSeeOther)
		return
	}

	if arch != "x86" && arch != "x64" {
		SendDebugMessage(fmt.Sprintf("❌ Invalid architecture: %s", arch))
		http.Redirect(w, r, "/?message=Invalid architecture&status=error", http.StatusSeeOther)
		return
	}

	// Save uploaded file
	SendDebugMessage(fmt.Sprintf("💾 Saving uploaded file: %s", handler.Filename))
	tmpShellcodePath, err := SaveUploadedFile(file, handler.Filename)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Error saving shellcode file: %v", err))
		http.Redirect(w, r, "/?message=Error saving shellcode file&status=error", http.StatusSeeOther)
		return
	}
	SendDebugMessage(fmt.Sprintf("✅ File saved to: %s", tmpShellcodePath))

	// Parse shellcode
	SendDebugMessage("🔍 Parsing shellcode...")
	cArray, err := TryAutoParseShellcode(tmpShellcodePath, selectedTemplate.Language)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Shellcode parse error: %v", err))
		http.Redirect(w, r, "/?message=Shellcode parse error&status=error", http.StatusSeeOther)
		return
	}
	SendDebugMessage(fmt.Sprintf("✅ Shellcode parsed successfully (%d bytes)", len(cArray)))

	// Generate random filename
	randomName := GenerateRandomFilename() + ".exe"
	outputPE := filepath.Join("output", randomName)
	SendDebugMessage(fmt.Sprintf("🎲 Generated output filename: %s", randomName))

	// Decide on encryption before processing template so placeholders are filled
	var encryptedProtocol string
	var encAlgoSelected string
	for k := range r.Form {
		if strings.HasSuffix(k, "_encryption") && r.FormValue(k) == "on" {
			encAlgoSelected = strings.ToUpper(strings.TrimSuffix(k, "_encryption"))
			break
		}
	}
	if encAlgoSelected == "" && selectedTemplate.SupportEncryption {
		encAlgoSelected = "AES"
	}

	var keyStr, ivStr, encShellcodeStr string
	if encAlgoSelected != "" {
		SendDebugMessage(fmt.Sprintf("🔐 Encryption required (%s). Running Supernova...", encAlgoSelected))
		k, v, es, errEnc := RunSupernovaEncryption(tmpShellcodePath, strings.Title(selectedTemplate.Language), encAlgoSelected)
		if errEnc != nil {
			SendDebugMessage(fmt.Sprintf("❌ %s encryption error: %v", encAlgoSelected, errEnc))
			http.Error(w, fmt.Sprintf("%s encryption error: %v", encAlgoSelected, errEnc), http.StatusInternalServerError)
			return
		}
		keyStr, ivStr, encShellcodeStr = k, v, es
		encryptedProtocol = "Encrypted-" + encAlgoSelected
	}

	// Prepare template data (with encryption fields if any)
	tplData := map[string]string{
		"shell_code":   cArray,
		"process_name": processName,
		"process_path": processPath,
	}
	if encAlgoSelected != "" {
		tplData["key"] = keyStr
		tplData["iv"] = ivStr
		tplData["encrypted_shellcode"] = encShellcodeStr
	}

	// Process template
	SendDebugMessage("📝 Processing template...")
	templateData, err := ProcessTemplate(selectedTemplate.Path, selectedTemplate.Language, tplData)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Template processing error: %v", err))
		http.Redirect(w, r, "/?message=Template processing error&status=error", http.StatusSeeOther)
		return
	}

	// Write source file
	var sourcePath string
	if selectedTemplate.Language == "c" {
		sourcePath = filepath.Join("output", "output.c")
	} else if selectedTemplate.Language == "csharp" {
		sourcePath = filepath.Join("output", "output.cs")
	} else if selectedTemplate.Language == "rust" {
		sourcePath = filepath.Join("output", "output.rs")
	}

	SendDebugMessage("💾 Writing source file...")
	err = WriteSourceFile(templateData, sourcePath)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to write source file: %v", err))
		http.Redirect(w, r, "/?message=Failed to write source file&status=error", http.StatusSeeOther)
		return
	}

	// Compile
	SendDebugMessage("🔨 Starting compilation...")
	var compileResult CompileResult
	if selectedTemplate.Language == "c" {
		hasEncryption := encAlgoSelected != ""
		compileResult = CompileC(sourcePath, outputPE, arch, hasEncryption)
	} else if selectedTemplate.Language == "csharp" {
		compileResult = CompileCSharp(sourcePath, outputPE, arch)
	} else if selectedTemplate.Language == "rust" {
		compileResult = CompileRust(selectedTemplate.Path, sourcePath, outputPE, arch)
	}

	if !compileResult.Success {
		SendDebugMessage("❌ Compilation failed")
		http.Redirect(w, r, "/?message=Compilation error&status=error", http.StatusSeeOther)
		return
	}

	// Encryption already handled before template processing

	// Run Astral-PE if requested
	if r.FormValue("ep_obfuscator") == "on" {
		SendDebugMessage("🔮 PE obfuscation requested with Astral-PE...")
		astralResult := RunAstralPE(outputPE, outputPE)
		if !astralResult.Success {
			SendDebugMessage(fmt.Sprintf("❌ Astral-PE obfuscation failed: %v", astralResult.Error))
			http.Error(w, "Error during PE obfuscation: "+astralResult.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Create metadata
	SendDebugMessage("📋 Creating metadata file...")
	metaPath := strings.TrimSuffix(outputPE, ".exe") + ".json"
	var obfuscated bool
	if r.FormValue("ep_obfuscator") == "on" {
		obfuscated = true
	}

	metaMap := map[string]interface{}{
		"filename":   filepath.Base(outputPE),
		"template":   templateName,
		"language":   selectedTemplate.Language,
		"arch":       arch,
		"created":    time.Now().Format("2006-01-02 15:04:05"),
		"obfuscated": obfuscated,
	}
	if encryptedProtocol != "" {
		metaMap["encrypted"] = encryptedProtocol
	}
	if processName != "" {
		metaMap["process_name"] = processName
	}

	metaBytes, _ := json.MarshalIndent(metaMap, "", "\t")
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Error writing metadata: %v", err))
		fmt.Printf("Error writing metadata: %v\n", err)
	} else {
		SendDebugMessage(fmt.Sprintf("✅ Metadata written to: %s", metaPath))
	}

	SendDebugMessage("🎉 Payload generation completed successfully!")
	SendDebugMessage(fmt.Sprintf("📁 Final output: %s", outputPE))
	if obfuscated {
		SendDebugMessage("🔮 PE obfuscation: Applied")
	}
	if encryptedProtocol != "" {
		SendDebugMessage(fmt.Sprintf("🔐 Encryption: %s", encryptedProtocol))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Payload generated successfully",
	})
}

// GetPayloads returns a list of all generated payloads
func GetPayloads() []Payload {
	var payloads []Payload
	files, err := os.ReadDir("output")
	if err != nil {
		fmt.Printf("Error reading output directory: %v\n", err)
		return payloads
	}

	payloadMap := make(map[string]Payload)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(strings.ToLower(filename), ".exe") {
			continue
		}

		baseName := strings.TrimSuffix(filename, ".exe")
		jsonPath := filepath.Join("output", baseName+".json")

		var payload Payload
		payload.Filename = filename

		if jsonData, err := os.ReadFile(jsonPath); err == nil {
			if err := json.Unmarshal(jsonData, &payload); err == nil {
				payload.Filename = filename
			}
		} else {
			payload = Payload{
				Filename:   filename,
				Created:    time.Now().Format("2006-01-02 15:04:05"),
				Template:   "Unknown",
				Language:   "Unknown",
				Arch:       "Unknown",
				Obfuscated: false,
				Encrypted:  "",
			}
		}

		payloadMap[filename] = payload
	}

	for _, payload := range payloadMap {
		payloads = append(payloads, payload)
	}

	sort.Slice(payloads, func(i, j int) bool {
		return payloads[i].Created > payloads[j].Created
	})

	return payloads
}

// DownloadsPageHandler handles the downloads page
func DownloadsPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("==> Rendering downloads page")
	payloads := GetPayloads()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Page":     "downloads",
		"Payloads": payloads,
	}
	if err := WebTemplates.ExecuteTemplate(w, "base.html", data); err != nil {
		fmt.Println("Template rendering error:", err)
		http.Error(w, "Template rendering error: "+err.Error(), http.StatusInternalServerError)
	}
}

// DeleteHandler handles payload deletion
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filename := filepath.Base(r.URL.Path)
	exePath := filepath.Join("output", filename)
	jsonPath := strings.TrimSuffix(exePath, ".exe") + ".json"
	os.Remove(exePath)
	os.Remove(jsonPath)
	w.WriteHeader(http.StatusOK)
}

// GetTemplateDetailsHandler handles template details API requests
func GetTemplateDetailsHandler(w http.ResponseWriter, r *http.Request) {
	queryTemplateName := r.URL.Query().Get("name")
	queryLanguage := r.URL.Query().Get("language")

	info, err := GetTemplateDetails(queryTemplateName, queryLanguage)
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// IndexPageHandler handles the main index page
func IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("message")
	status := r.URL.Query().Get("status")
	data := map[string]interface{}{
		"Page":      "index",
		"Message":   message,
		"Status":    status,
		"Templates": AvailableTemplates,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := WebTemplates.ExecuteTemplate(w, "base.html", data); err != nil {
		fmt.Println("Template error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
