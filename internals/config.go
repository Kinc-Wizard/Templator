package internals

import (
	"encoding/json"
	"fmt"
	"os"
)

// AppConfig represents the application configuration
type AppConfig struct {
	MinGWx86LibPath   string `json:"mingw_x86_lib_path"`
	MinGWx64LibPath   string `json:"mingw_x64_lib_path"`
	MonoFrameworkPath string `json:"mono_framework_path"`
	AstralPEPath      string `json:"astral_pe_path"`
	CCompilerX86      string `json:"c_compiler_x86"`
	CCompilerX64      string `json:"c_compiler_x64"`
	CSharpCompiler    string `json:"csharp_compiler"`
}

// Global variable to hold the loaded config
var Config AppConfig

// LoadAppConfig reads the configuration from config.json (optional)
func LoadAppConfig() {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Printf("⚠️ Could not read config.json: %v. Proceeding with defaults.\n", err)
		return
	}

	err = json.Unmarshal(configFile, &Config)
	if err != nil {
		fmt.Printf("⚠️ Could not parse config.json: %v. Proceeding with defaults.\n", err)
		return
	}
	// Configuration loaded successfully (optional)
}
