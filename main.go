package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"templator/internals"
)

func main() {
	// Load configuration at startup (optional)
	if _, err := os.Stat("config.json"); err == nil {
		internals.LoadAppConfig()
	} else {
		internals.SendDebugMessage("⚠️ config.json not found; using defaults and PATH tools")
	}

	// Create necessary directories
	os.MkdirAll("output", 0755)
	os.MkdirAll("uploads", 0755)

	// Initialize web templates
	internals.InitWebTemplates()

	// Scan templates only once at startup
	internals.AvailableTemplates = internals.ScanAvailableTemplates()

	// Announce loaded templates concisely in the web UI terminal
	if len(internals.AvailableTemplates) > 0 {
		langToNames := make(map[string][]string)
		for _, t := range internals.AvailableTemplates {
			langToNames[t.Language] = append(langToNames[t.Language], t.Name)
		}
		internals.SendDebugMessage(fmt.Sprintf("✅ %d shellcode template(s) loaded", len(internals.AvailableTemplates)))
		var langs []string
		for lang := range langToNames {
			langs = append(langs, lang)
		}
		sort.Strings(langs)
		for _, lang := range langs {
			names := langToNames[lang]
			sort.Strings(names)
			internals.SendDebugMessage(fmt.Sprintf("📦 %s: %s", lang, strings.Join(names, ", ")))
		}
	} else {
		internals.SendDebugMessage("⚠️ No shellcode templates found")
	}

	// Setup HTTP routes
	http.HandleFunc("/", internals.IndexPageHandler)
	http.HandleFunc("/upload", internals.UploadHandler)
	http.HandleFunc("/downloads", internals.DownloadsPageHandler)
	http.HandleFunc("/delete/", internals.DeleteHandler)
	http.HandleFunc("/terminal_ws", internals.TerminalWSHandler)
	http.HandleFunc("/ws/terminal", internals.TerminalWSHandler)
	http.HandleFunc("/template_details", internals.GetTemplateDetailsHandler)
	http.HandleFunc("/template/details", internals.GetTemplateDetailsHandler)

	// Static file servers
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/output/", http.StripPrefix("/output/", http.FileServer(http.Dir("output"))))

	fmt.Println("http://localhost:12345")
	http.ListenAndServe(":12345", nil)
}
