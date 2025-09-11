package internals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Template represents a shellcode template
type Template struct {
	Name                string
	Path                string
	Language            string
	Description         string
	Details             string
	NeedsProcess        bool
	SupportsObfuscation bool
	SupportEncryption   bool
	NeedsProcessPath    bool
	Compile             map[string]string
}

// TemplateInfo represents template information for API responses
type TemplateInfo struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	Language            string            `json:"language"`
	SupportsObfuscation bool              `json:"supports_obfuscation"`
	NeedsProcess        bool              `json:"needs_process"`
	SupportEncryption   bool              `json:"support_encryption"`
	NeedsProcessPath    bool              `json:"needs_process_path"`
	Compile             map[string]string `json:"compile"`
}

// Global variable to store templates
var AvailableTemplates []Template

// ScanAvailableTemplates scans the templates directory and returns available templates
func ScanAvailableTemplates() []Template {
	var templates []Template
	templatesDir := "templates_shellcode"

	// Read templates.json
	templatesFile, err := os.ReadFile(filepath.Join(templatesDir, "templates.json"))
	if err != nil {
		fmt.Printf("Error reading templates.json: %v\n", err)
		return templates
	}

	var templateInfos map[string]TemplateInfo
	if err := json.Unmarshal(templatesFile, &templateInfos); err != nil {
		fmt.Printf("Error parsing templates.json: %v\n", err)
		return templates
	}

	// Map to track unique templates
	seenTemplates := make(map[string]bool)

	// Scan all language folders
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		fmt.Printf("Error reading templates directory: %v\n", err)
		return templates
	}

	// Process each language folder
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." {
			langPath := filepath.Join(templatesDir, entry.Name())
			language := entry.Name()

			// Handle special case for Rust which has subdirectories
			if language == "rust" {
				// Read all subdirectories in the rust folder
				rustDirs, err := os.ReadDir(langPath)
				if err != nil {
					fmt.Printf("Error reading rust directory: %v\n", err)
					continue
				}

				for _, rustDir := range rustDirs {
					if rustDir.IsDir() {
						// Look for the main Rust source file
						rustSrcPath := filepath.Join(langPath, rustDir.Name(), "src")
						rustFiles, err := os.ReadDir(rustSrcPath)
						if err != nil {
							continue
						}

						for _, rustFile := range rustFiles {
							if !rustFile.IsDir() && strings.HasSuffix(rustFile.Name(), ".rs") {
								templateName := strings.TrimSuffix(rustFile.Name(), ".rs")
								templatePath := filepath.Join(rustSrcPath, rustFile.Name())

								// Check if we've already seen this template for this language
								uniqueKey := language + "/" + templateName
								if !seenTemplates[uniqueKey] {
									seenTemplates[uniqueKey] = true

									// Lookup info by base name first, then fallback to full filename
									info, exists := templateInfos[templateName]
									if !exists {
										info, exists = templateInfos[rustFile.Name()]
									}
									if exists {
										templates = append(templates, Template{
											Name:                templateName,
											Path:                templatePath,
											Language:            language,
											Description:         info.Description,
											SupportsObfuscation: info.SupportsObfuscation,
											NeedsProcess:        info.NeedsProcess,
											SupportEncryption:   info.SupportEncryption,
											NeedsProcessPath:    info.NeedsProcessPath,
											Compile:             info.Compile,
										})
									} else {
										templates = append(templates, Template{
											Name:                templateName,
											Path:                templatePath,
											Language:            language,
											Description:         templateName,
											SupportsObfuscation: false,
											NeedsProcess:        false,
											SupportEncryption:   false,
											NeedsProcessPath:    false,
											Compile:             nil,
										})
									}
								}
							}
						}
					}
				}
				continue
			}

			// Handle regular directories
			files, err := os.ReadDir(langPath)
			if err != nil {
				fmt.Printf("Error reading %s directory: %v\n", language, err)
				continue
			}

			// Process each template file
			for _, file := range files {
				if !file.IsDir() {
					templateName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
					templatePath := filepath.Join(langPath, file.Name())

					// Check if we've already seen this template for this language
					uniqueKey := language + "/" + templateName
					if !seenTemplates[uniqueKey] {
						seenTemplates[uniqueKey] = true

						// Lookup info by base name first, then fallback to full filename
						info, exists := templateInfos[templateName]
						if !exists {
							info, exists = templateInfos[file.Name()]
						}
						if exists {
							templates = append(templates, Template{
								Name:                templateName,
								Path:                templatePath,
								Language:            language,
								Description:         info.Description,
								SupportsObfuscation: info.SupportsObfuscation,
								NeedsProcess:        info.NeedsProcess,
								SupportEncryption:   info.SupportEncryption,
								NeedsProcessPath:    info.NeedsProcessPath,
								Compile:             info.Compile,
							})
						} else {
							templates = append(templates, Template{
								Name:                templateName,
								Path:                templatePath,
								Language:            language,
								Description:         templateName,
								SupportsObfuscation: false,
								NeedsProcess:        false,
								SupportEncryption:   false,
								NeedsProcessPath:    false,
								Compile:             nil,
							})
						}
					}
				}
			}
		}
	}

	// Sort templates by language and name for consistent display
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].Language != templates[j].Language {
			return templates[i].Language < templates[j].Language
		}
		return templates[i].Name < templates[j].Name
	})

	fmt.Printf("Found %d unique templates across all languages\n", len(templates))
	return templates
}

// GetTemplateDetails returns details for a specific template
func GetTemplateDetails(templateName, language string) (*TemplateInfo, error) {
	for _, tpl := range AvailableTemplates {
		if tpl.Name == templateName && tpl.Language == language {
			info := &TemplateInfo{
				Name:                tpl.Name,
				Description:         tpl.Description,
				Language:            tpl.Language,
				SupportsObfuscation: tpl.SupportsObfuscation,
				NeedsProcess:        tpl.NeedsProcess,
				SupportEncryption:   tpl.SupportEncryption,
				NeedsProcessPath:    tpl.NeedsProcessPath,
				Compile:             tpl.Compile,
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("template not found")
}

// ParseTemplate replaces placeholders in template content
func ParseTemplate(templateContent string, data map[string]string) (string, error) {
	// Replace placeholders one by one
	result := templateContent

	// Replace main placeholders
	if shellCode, ok := data["shell_code"]; ok {
		result = strings.ReplaceAll(result, "{{shell_code}}", shellCode)
	}
	if processName, ok := data["process_name"]; ok && processName != "" {
		result = strings.ReplaceAll(result, "{{process_name}}", processName)
	}
	if processPath, ok := data["process_path"]; ok && processPath != "" {
		result = strings.ReplaceAll(result, "{{process_path}}", processPath)
	}
	if key, ok := data["key"]; ok && key != "" {
		result = strings.Replace(result, "{{key}}", key, 1)
	}
	if iv, ok := data["iv"]; ok && iv != "" {
		result = strings.Replace(result, "{{iv}}", iv, 1)
	}
	if encShellcode, ok := data["encrypted_shellcode"]; ok && encShellcode != "" {
		result = strings.Replace(result, "{{encrypted_shellcode}}", encShellcode, 1)
	}

	return result, nil
}
