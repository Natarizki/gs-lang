package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gs-lang/bytecode"
	"gs-lang/lexer"
	"gs-lang/parser"
	"gs-lang/vm"
)

func main() {
	binName := filepath.Base(os.Args[0])

	if binName == "gsi" {
		if len(os.Args) < 2 {
			fmt.Println("Usage: gsi <username/paket>")
			os.Exit(1)
		}
		installPackage(os.Args[1])
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gsc run <file.gs>")
			os.Exit(1)
		}
		runFile(os.Args[2])
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gsc build <file.gs> -o <output>")
			os.Exit(1)
		}
		sourceFile := os.Args[2]
		outputFile := "a.out"
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "-o" && i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
			}
		}
		buildBinary(sourceFile, outputFile)
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gsc install <username/paket>")
			os.Exit(1)
		}
		installPackage(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GS (GoSong) - CLI")
	fmt.Println("Usage:")
	fmt.Println("  gsc run <file.gs>   Run a GS source file directly")
	fmt.Println("  gsc build           Build the current GS project into a binary")
}

func runFile(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", path, err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	program := p.ParseProgramStrict()

	if len(p.Errors()) > 0 {
		fmt.Println("Parse errors:")
		for _, e := range p.Errors() {
			fmt.Println("  " + e)
		}
		os.Exit(1)
	}

        program, err = resolveImports(program, filepath.Dir(path), make(map[string]bool))
	if err != nil {
		fmt.Printf("Import error: %s\n", err)
		os.Exit(1)
	}

	gen := bytecode.New()
	if err := gen.Compile(program); err != nil {
		fmt.Printf("Compile error: %s\n", err)
		os.Exit(1)
	}

	bc := gen.Bytecode()
	if os.Getenv("GSC_DEBUG") != "" {
		fmt.Println("=== BYTECODE ===")
		fmt.Println(bc.Instructions.String())
		for i, c := range bc.Constants {
			if cf, ok := c.(*bytecode.CompiledFunctionConstant); ok {
				fmt.Printf("=== CONST %d (function) ===\n", i)
				fmt.Println(cf.Instructions.String())
			}
		}
		fmt.Println("=== END BYTECODE ===")
	}

	machine := vm.New(bc)
	if err := machine.Run(); err != nil {
		fmt.Printf("Runtime error: %s\n", err)
		os.Exit(1)
	}
}

// resolveImports menelusuri semua ImportStatement dalam sebuah Program,
// membaca file .gs yang dituju (relatif terhadap baseDir), lalu menggabungkan
// isi file tersebut ke dalam Program utama. visited mencegah import berulang
// (circular import) dari file yang sama.
func resolveImports(program *parser.Program, baseDir string, visited map[string]bool) (*parser.Program, error) {
	newStatements := []parser.Statement{}

	for _, stmt := range program.Statements {
		importStmt, ok := stmt.(*parser.ImportStatement)
		if !ok {
			newStatements = append(newStatements, stmt)
			continue
		}

		if importStmt.Path == "GS/af" || importStmt.Path == "GS/fo" {
			newStatements = append(newStatements, stmt)
			continue
		}

		var filesToLoad []string

		if strings.HasPrefix(importStmt.Path, "GS/") {
			pkgPath := strings.TrimPrefix(importStmt.Path, "GS/")
			pkgDir, err := resolvePackageDir(pkgPath)
			if err != nil {
				return nil, err
			}
			entries, err := os.ReadDir(pkgDir)
			if err != nil {
				return nil, fmt.Errorf("cannot read package directory %q: %s", pkgDir, err)
			}
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".gs") {
					filesToLoad = append(filesToLoad, filepath.Join(pkgDir, e.Name()))
				}
			}
		} else {
			importPath := filepath.Join(baseDir, importStmt.Path)
			absPath, err := filepath.Abs(importPath)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve import path %q: %s", importStmt.Path, err)
			}
			filesToLoad = append(filesToLoad, absPath)
		}

		for _, absPath := range filesToLoad {
			if visited[absPath] {
				continue // sudah pernah di-import, skip (cegah circular import)
			}
			visited[absPath] = true

			src, err := os.ReadFile(absPath)
			if err != nil {
				return nil, fmt.Errorf("cannot read imported file %q: %s", absPath, err)
			}

			l := lexer.New(string(src))
			p := parser.New(l)
			importedProgram := p.ParseProgram()

			if len(p.Errors()) > 0 {
				msg := fmt.Sprintf("parse errors in imported file %q:", absPath)
				for _, e := range p.Errors() {
					msg += "\n  " + e
				}
				return nil, fmt.Errorf("%s", msg)
			}

			importedProgram, err = resolveImports(importedProgram, filepath.Dir(absPath), visited)
			if err != nil {
				return nil, err
			}

			newStatements = append(newStatements, importedProgram.Statements...)
		}
	}

	return &parser.Program{Statements: newStatements}, nil
}

// installPackage mendownload package GS dari GitHub (username/paket) ke ~/.GS/username/paket/
func installPackage(pkgPath string) {
	parts := strings.Split(pkgPath, "/")
	if len(parts) != 2 {
		fmt.Printf("Invalid package format: %s (expected username/paket)\n", pkgPath)
		os.Exit(1)
	}
	username, pkgName := parts[0], parts[1]

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error finding home directory: %s\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".GS", username, pkgName)

	if _, err := os.Stat(destDir); err == nil {
		fmt.Printf("Package %s already installed at %s\n", pkgPath, destDir)
		return
	}

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", username, pkgName)
	fmt.Printf("Installing %s from %s...\n", pkgPath, repoURL)

	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		fmt.Printf("Error creating install directory: %s\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error installing package: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Installed %s to %s\n", pkgPath, destDir)
}

// resolvePackageDir mengembalikan path folder package di ~/.GS/username/paket,
// auto-install dulu kalau belum ada
func resolvePackageDir(pkgPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %s", err)
	}

	destDir := filepath.Join(homeDir, ".GS", pkgPath)

	if _, err := os.Stat(destDir); err == nil {
		return destDir, nil
	}

	parts := strings.Split(pkgPath, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid package path %q (expected username/paket)", pkgPath)
	}
	username, pkgName := parts[0], parts[1]

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", username, pkgName)
	fmt.Printf("Auto-installing GS/%s from %s...\n", pkgPath, repoURL)

	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return "", fmt.Errorf("cannot create install directory: %s", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cannot install package %s: %s", pkgPath, err)
	}

	return destDir, nil
}

// buildBinary meng-compile file .gs menjadi bytecode, lalu membungkusnya
// menjadi binary Go standalone (embed VM + bytecode) di path output yang diberikan
func buildBinary(sourcePath string, outputPath string) {
	src, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", sourcePath, err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	program := p.ParseProgramStrict()

	if len(p.Errors()) > 0 {
		fmt.Println("Parse errors:")
		for _, e := range p.Errors() {
			fmt.Println("  " + e)
		}
		os.Exit(1)
	}

	program, err = resolveImports(program, filepath.Dir(sourcePath), make(map[string]bool))
	if err != nil {
		fmt.Printf("Import error: %s\n", err)
		os.Exit(1)
	}

	gen := bytecode.New()
	if err := gen.Compile(program); err != nil {
		fmt.Printf("Compile error: %s\n", err)
		os.Exit(1)
	}

	bc := gen.Bytecode()

	tmpDir, err := os.MkdirTemp("", "gsc-build-*")
	if err != nil {
		fmt.Printf("Error creating temp directory: %s\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	gobPath := filepath.Join(tmpDir, "bytecode.gob")
	if err := saveBytecodeGob(bc, gobPath); err != nil {
		fmt.Printf("Error serializing bytecode: %s\n", err)
		os.Exit(1)
	}

	if err := generateBuildMain(tmpDir, gobPath); err != nil {
		fmt.Printf("Error generating build source: %s\n", err)
		os.Exit(1)
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		fmt.Printf("Error resolving output path: %s\n", err)
		os.Exit(1)
	}

       tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		fmt.Printf("Error running go mod tidy: %s\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", absOutput, ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error building binary: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built %s\n", absOutput)
}

// saveBytecodeGob menyerialisasi bc.Instructions dan bc.Constants ke file
// menggunakan encoding/gob, dengan registrasi tipe konkret yang dipakai
// di dalam constant pool
func saveBytecodeGob(bc *bytecode.Bytecode, path string) error {
	gob.Register(int64(0))
	gob.Register(float64(0))
	gob.Register("")
	gob.Register([]string{})
	gob.Register(&bytecode.CompiledFunctionConstant{})
	gob.Register(&bytecode.StructDefConstant{})

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	payload := struct {
		Instructions bytecode.Instructions
		Constants    []interface{}
	}{
		Instructions: bc.Instructions,
		Constants:    bc.Constants,
	}
	return enc.Encode(payload)
}

// generateBuildMain membuat project Go sementara di tmpDir yang meng-embed
// file bytecode.gob dan menjalankannya lewat VM saat binary dieksekusi
func generateBuildMain(tmpDir string, gobPath string) error {

	gsLangPath, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	goModContent := "module gsc_build_output\n\n" +
		"go 1.21\n\n" +
		"require gs-lang v0.0.0\n\n" +
		"replace gs-lang => " + gsLangPath + "\n"

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}

	gobData, err := os.ReadFile(gobPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "bytecode.gob"), gobData, 0644); err != nil {
		return err
	}

	mainContent := `package main

import (
	_ "embed"
	"encoding/gob"
	"bytes"
	"fmt"
	"os"

	"gs-lang/bytecode"
	"gs-lang/vm"
)

//go:embed bytecode.gob
var embeddedBytecode []byte

func main() {
	gob.Register(int64(0))
	gob.Register(float64(0))
	gob.Register("")
	gob.Register([]string{})
	gob.Register(&bytecode.CompiledFunctionConstant{})
	gob.Register(&bytecode.StructDefConstant{})

	var payload struct {
		Instructions bytecode.Instructions
		Constants    []interface{}
	}

	dec := gob.NewDecoder(bytes.NewReader(embeddedBytecode))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("Error loading embedded bytecode: %s\n", err)
		os.Exit(1)
	}

	bc := &bytecode.Bytecode{Instructions: payload.Instructions, Constants: payload.Constants}
	machine := vm.New(bc)
	if err := machine.Run(); err != nil {
		fmt.Printf("Runtime error: %s\n", err)
		os.Exit(1)
	}
}
`
	return os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainContent), 0644)
}
