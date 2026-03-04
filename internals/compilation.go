package internals

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// CompileResult represents the result of a compilation
type CompileResult struct {
	Success bool
	Output  []byte
	Error   error
}

// CompileC compiles C source code
func CompileC(sourcePath, outputPath, arch string, hasEncryption bool) CompileResult {
	SendDebugMessage(fmt.Sprintf("🔨 Compiling C source for %s architecture...", arch))
	
	var compiler string
	if arch == "x86" {
		if Config.CCompilerX86 != "" { compiler = Config.CCompilerX86 } else { compiler = "i686-w64-mingw32-gcc" }
	} else {
		if Config.CCompilerX64 != "" { compiler = Config.CCompilerX64 } else { compiler = "x86_64-w64-mingw32-gcc" }
	}

	var libPath string
	if arch == "x86" {
		libPath = Config.MinGWx86LibPath
	} else {
		libPath = Config.MinGWx64LibPath
	}

	// Add -lbcrypt at the very end of the arguments (important for linking under mingw)
	compileArgs := []string{sourcePath, "-o", outputPath, "-fno-pie", "-static"}
	if libPath != "" {
		compileArgs = append(compileArgs, "-L"+libPath)
	}
	if hasEncryption {
		SendDebugMessage("🔐 Adding encryption libraries (bcrypt)...")
		compileArgs = append(compileArgs, "-Wl,--start-group", "-lbcrypt", "-Wl,--end-group")
	}

	SendDebugMessage(fmt.Sprintf("⚙️ Using compiler: %s", compiler))
	SendDebugMessage(fmt.Sprintf("📁 Output: %s", outputPath))

	cmd := exec.Command(compiler, compileArgs...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ C compilation failed to start: %v", err))
		return CompileResult{Success: false, Output: nil, Error: err}
	}
	var buf strings.Builder
	var wg sync.WaitGroup
	stream := func(r io.Reader, prefix string) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			buf.WriteString(line + "\n")
			SendDebugMessage(prefix + line)
		}
	}
	wg.Add(2)
	go stream(stdout, "📤 ")
	go stream(stderr, "⚠️ ")
	err := cmd.Wait()
	wg.Wait()
	output := []byte(buf.String())

	if err == nil {
		SendDebugMessage("✅ C compilation successful!")
	} else {
		SendDebugMessage(fmt.Sprintf("❌ C compilation failed: %v", err))
	}
	
	return CompileResult{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// CompileCSharp compiles C# source code
func CompileCSharp(sourcePath, outputPath, arch string) CompileResult {
	SendDebugMessage(fmt.Sprintf("🔨 Compiling C# source for %s architecture...", arch))
	
	// Prefer standard compilers if config is empty
	csCompilerComponents := strings.Fields(strings.TrimSpace(Config.CSharpCompiler))
	csCompilerExe := "mcs"
	csCompilerArgs := []string{}
	if len(csCompilerComponents) > 0 && csCompilerComponents[0] != "" {
		csCompilerExe = csCompilerComponents[0]
		if len(csCompilerComponents) > 1 {
			csCompilerArgs = append(csCompilerArgs, csCompilerComponents[1:]...)
		}
	}

	monoPath := strings.TrimSpace(Config.MonoFrameworkPath)
	dllPath := ""
	if monoPath != "" {
		dllPath = filepath.Join(monoPath, "Facades", "System.Runtime.InteropServices.dll")
	}

	csCompilerArgs = append(csCompilerArgs,
		"-target:winexe",
		"-platform:"+strings.ToLower(arch),
		"-unsafe",
	)
	if monoPath != "" {
		csCompilerArgs = append(csCompilerArgs,
			fmt.Sprintf("-lib:%s", monoPath),
			fmt.Sprintf("-r:%s", dllPath),
		)
	}
	csCompilerArgs = append(csCompilerArgs,
		"-out:"+outputPath,
		sourcePath,
	)

	SendDebugMessage(fmt.Sprintf("⚙️ Using C# compiler: %s", csCompilerExe))
	SendDebugMessage(fmt.Sprintf("📁 Output: %s", outputPath))
	if monoPath != "" { SendDebugMessage(fmt.Sprintf("📚 Mono framework: %s", monoPath)) }

	compileCmd := exec.Command(csCompilerExe, csCompilerArgs...)
	stdout, _ := compileCmd.StdoutPipe()
	stderr, _ := compileCmd.StderrPipe()
	if err := compileCmd.Start(); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ C# compilation failed to start: %v", err))
		return CompileResult{Success: false, Output: nil, Error: err}
	}
	var buf2 strings.Builder
	var wg2 sync.WaitGroup
	stream := func(r io.Reader, prefix string) {
		defer wg2.Done()
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			buf2.WriteString(line + "\n")
			SendDebugMessage(prefix + line)
		}
	}
	wg2.Add(2)
	go stream(stdout, "📤 ")
	go stream(stderr, "⚠️ ")
	err := compileCmd.Wait()
	wg2.Wait()
	output := []byte(buf2.String())

	if err == nil {
		SendDebugMessage("✅ C# compilation successful!")
	} else {
		SendDebugMessage(fmt.Sprintf("❌ C# compilation failed: %v", err))
	}
	
	return CompileResult{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// ProcessTemplate processes a template with the given data
func ProcessTemplate(templatePath, language string, data map[string]string) (string, error) {
	SendDebugMessage(fmt.Sprintf("📝 Processing %s template: %s", language, templatePath))
	
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to read template: %v", err))
		return "", fmt.Errorf("failed to read template: %v", err)
	}

	SendDebugMessage(fmt.Sprintf("📖 Template file read successfully (%d bytes)", len(templateData)))

	if language == "c" {
		SendDebugMessage("🔧 Processing C template placeholders...")
		fullCode := string(templateData)
		
		// Replace placeholders
		if key, ok := data["key"]; ok && key != "" {
			SendDebugMessage("🔑 Replacing encryption key placeholder")
			fullCode = strings.Replace(fullCode, "{{key}}", key, 1)
		}
		if iv, ok := data["iv"]; ok && iv != "" {
			SendDebugMessage("🔐 Replacing IV placeholder")
			fullCode = strings.Replace(fullCode, "{{iv}}", iv, 1)
		}
		if encShellcode, ok := data["encrypted_shellcode"]; ok && encShellcode != "" {
			SendDebugMessage("🔒 Replacing encrypted shellcode placeholder")
			fullCode = strings.Replace(fullCode, "{{encrypted_shellcode}}", encShellcode, 1)
		} else if shellCode, ok := data["shell_code"]; ok {
			SendDebugMessage("💉 Replacing shellcode placeholder")
			fullCode = strings.Replace(fullCode, "{{shell_code}}", shellCode, 1)
		}
		
		// Always replace process_name if provided
		if processName, ok := data["process_name"]; ok && processName != "" {
			SendDebugMessage(fmt.Sprintf("🎯 Replacing process name: %s", processName))
			fullCode = strings.ReplaceAll(fullCode, "{{process_name}}", processName)
		}
		// Replace process_path if needed
		if processPath, ok := data["process_path"]; ok && processPath != "" {
			SendDebugMessage(fmt.Sprintf("📂 Replacing process path: %s", processPath))
			fullCode = strings.ReplaceAll(fullCode, "{{process_path}}", processPath)
		}
		
		SendDebugMessage("✅ C template processing completed")
		return fullCode, nil
	} else if language == "csharp" {
		SendDebugMessage("🔧 Processing C# template with advanced parser...")
		result, err := ParseTemplate(string(templateData), data)
		if err != nil {
			SendDebugMessage(fmt.Sprintf("❌ C# template processing failed: %v", err))
		} else {
			SendDebugMessage("✅ C# template processing completed")
		}
		return result, err
	} else if language == "rust" {
		SendDebugMessage("🔧 Processing Rust template with advanced parser...")
		result, err := ParseTemplate(string(templateData), data)
		if err != nil {
			SendDebugMessage(fmt.Sprintf("❌ Rust template processing failed: %v", err))
		} else {
			SendDebugMessage("✅ Rust template processing completed")
		}
		return result, err
	}

	SendDebugMessage(fmt.Sprintf("❌ Unsupported language: %s", language))
	return "", fmt.Errorf("unsupported language: %s", language)
}

// CompileRust compiles Rust source code using rustc targeting Windows (gnu toolchain)
func CompileRust(templatePath, sourcePath, outputPath, arch string) CompileResult {
    SendDebugMessage(fmt.Sprintf("🔨 Compiling Rust source for %s architecture...", arch))

    // Determine target triple
    var target string
    if arch == "x86" {
        target = "i686-pc-windows-gnu"
    } else {
        target = "x86_64-pc-windows-gnu"
    }

    // Reference project directory (templates_shellcode/rust/<name>)
    projectDir := filepath.Dir(filepath.Dir(templatePath))

    // Prepare temporary cargo project under output/
    tempDir := filepath.Join("output", "rust_build_"+GenerateRandomFilename())
    srcDir := filepath.Join(tempDir, "src")
    if err := os.MkdirAll(srcDir, 0755); err != nil {
        return CompileResult{Success: false, Output: nil, Error: fmt.Errorf("failed to create temp rust dir: %w", err)}
    }

    // Copy Cargo.toml from the reference template
    cargoSrc := filepath.Join(projectDir, "Cargo.toml")
    cargoDst := filepath.Join(tempDir, "Cargo.toml")
    cargoBytes, err := os.ReadFile(cargoSrc)
    if err != nil {
        return CompileResult{Success: false, Output: nil, Error: fmt.Errorf("failed to read Cargo.toml: %w", err)}
    }
    if err := os.WriteFile(cargoDst, cargoBytes, 0644); err != nil {
        return CompileResult{Success: false, Output: nil, Error: fmt.Errorf("failed to write Cargo.toml: %w", err)}
    }

    // Read processed source and write as main.rs
    mainBytes, err := os.ReadFile(sourcePath)
    if err != nil {
        return CompileResult{Success: false, Output: nil, Error: fmt.Errorf("failed to read processed rust source: %w", err)}
    }
    mainDst := filepath.Join(srcDir, "main.rs")
    if err := os.WriteFile(mainDst, mainBytes, 0644); err != nil {
        return CompileResult{Success: false, Output: nil, Error: fmt.Errorf("failed to write main.rs: %w", err)}
    }

    // Determine binary name from Cargo.toml (fallback to folder name)
    binName := filepath.Base(projectDir)
    lines := strings.Split(string(cargoBytes), "\n")
    for _, line := range lines {
        lineTrim := strings.TrimSpace(line)
        if strings.HasPrefix(lineTrim, "name = ") {
            parts := strings.SplitN(lineTrim, "\"", 3)
            if len(parts) >= 2 {
                binName = parts[1]
            }
            break
        }
    }

    // Build with cargo
    SendDebugMessage("⚙️ Using Cargo to build Rust project")
    SendDebugMessage(fmt.Sprintf("📁 Temp project: %s", tempDir))
    SendDebugMessage(fmt.Sprintf("🎯 Target: %s", target))
    cmd := exec.Command("cargo", "build", "--release", "--target", target, "--quiet")
    cmd.Dir = tempDir
    // Further reduce verbosity: silence warnings and disable color
    cmd.Env = append(os.Environ(),
        "RUSTFLAGS=-Awarnings",
        "CARGO_TERM_COLOR=never",
    )
    output, err := cmd.CombinedOutput()
    if err != nil {
        SendDebugMessage(fmt.Sprintf("❌ Rust cargo build failed: %v", err))
        SendDebugMessage(fmt.Sprintf("📄 Cargo output: %s", string(output)))
        return CompileResult{Success: false, Output: output, Error: err}
    }

    // Copy produced .exe to desired output
    builtExe := filepath.Join(tempDir, "target", target, "release", binName+".exe")
    exeBytes, err := os.ReadFile(builtExe)
    if err != nil {
        SendDebugMessage(fmt.Sprintf("❌ Failed to read built exe: %v", err))
        return CompileResult{Success: false, Output: output, Error: err}
    }
    if err := os.WriteFile(outputPath, exeBytes, 0644); err != nil {
        SendDebugMessage(fmt.Sprintf("❌ Failed to write final exe: %v", err))
        return CompileResult{Success: false, Output: output, Error: err}
    }

    SendDebugMessage("✅ Rust compilation successful!")
    SendDebugMessage(fmt.Sprintf("📦 Output file: %s", outputPath))
    return CompileResult{Success: true, Output: output, Error: nil}
}

// WriteSourceFile writes processed template to a source file
func WriteSourceFile(content, outputPath string) error {
	SendDebugMessage(fmt.Sprintf("💾 Writing source file: %s", outputPath))
	SendDebugMessage(fmt.Sprintf("📊 File size: %d bytes", len(content)))
	
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to write source file: %v", err))
	} else {
		SendDebugMessage("✅ Source file written successfully")
	}
	
	return err
}

// RunAstralPE runs Astral-PE obfuscation on the output file
func RunAstralPE(inputPath, outputPath string) CompileResult {
	SendDebugMessage("🔮 Starting Astral-PE obfuscation...")
	SendDebugMessage(fmt.Sprintf("📁 Input: %s", inputPath))
	SendDebugMessage(fmt.Sprintf("📁 Output: %s", outputPath))
	
	astralPath := Config.AstralPEPath
	if strings.TrimSpace(astralPath) == "" {
		astralPath = filepath.Join("tools", "native", "Astral-PE")
	}

	if _, err := os.Stat(astralPath); os.IsNotExist(err) {
		SendDebugMessage(fmt.Sprintf("❌ Astral-PE not found at: %s", astralPath))
		return CompileResult{
			Success: false,
			Output:  []byte(fmt.Sprintf("Astral-PE not found at configured path: %s", astralPath)),
			Error:   fmt.Errorf("astral-pe not found at: %s", astralPath),
		}
	}

	SendDebugMessage(fmt.Sprintf("⚙️ Using Astral-PE: %s", astralPath))

	// Astral-PE will overwrite the original file
	cmd := exec.Command(astralPath, inputPath, "-o", outputPath)
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		SendDebugMessage("✅ Astral-PE obfuscation successful!")
	} else {
		SendDebugMessage(fmt.Sprintf("❌ Astral-PE obfuscation failed: %v", err))
		SendDebugMessage(fmt.Sprintf("📄 Astral-PE output: %s", string(output)))
	}
	
	return CompileResult{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// RunCustomCompile executes a user-defined compile command with live streaming
func RunCustomCompile(cmdTemplate string, placeholders map[string]string, workDir string) CompileResult {
	cmdStr := cmdTemplate
	for k, v := range placeholders {
		cmdStr = strings.ReplaceAll(cmdStr, "{"+k+"}", v)
	}
	SendDebugMessage("🔨 Custom compile started")
	SendDebugMessage("🧱 Command: " + cmdStr)
	cmd := exec.Command("/bin/sh", "-lc", cmdStr)
	if workDir != "" {
		cmd.Dir = workDir
	}
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		SendDebugMessage(fmt.Sprintf("❌ Failed to start command: %v", err))
		return CompileResult{Success: false, Output: nil, Error: err}
	}
	var buf strings.Builder
	var wg sync.WaitGroup
	stream := func(r io.Reader, prefix string) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			buf.WriteString(line + "\n")
			SendDebugMessage(prefix + line)
		}
	}
	wg.Add(2)
	go stream(stdout, "📤 ")
	go stream(stderr, "⚠️ ")
	err := cmd.Wait()
	wg.Wait()
	outBytes := []byte(buf.String())
	if err != nil {
		SendDebugMessage("❌ Custom compile failed")
		return CompileResult{Success: false, Output: outBytes, Error: err}
	}
	SendDebugMessage("✅ Custom compile finished")
	return CompileResult{Success: true, Output: outBytes, Error: nil}
}
