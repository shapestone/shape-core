package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
	"github.com/shapestone/shape/pkg/shape"
	"github.com/shapestone/shape/pkg/validator"
)

var (
	formatFlag    = flag.String("f", "auto", "Schema format (jsonv, xmlv, yamlv, csvv, propsv, textv, auto)")
	outputFlag    = flag.String("o", "text", "Output format (text, json, quiet)")
	noColorFlag   = flag.Bool("no-color", false, "Disable colored output")
	registerTypes = flag.String("register-type", "", "Register custom types (comma-separated)")
	verboseFlag   = flag.Bool("v", false, "Verbose output")
	versionFlag   = flag.Bool("version", false, "Show version")
)

const version = "0.3.0"

func main() {
	flag.Usage = usage
	flag.Parse()

	// Show version and exit
	if *versionFlag {
		fmt.Printf("shape-validate version %s\n", version)
		os.Exit(0)
	}

	// Check arguments
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: No schema files specified")
		flag.Usage()
		os.Exit(3)
	}

	// Set NO_COLOR if --no-color flag is set
	if *noColorFlag {
		os.Setenv("NO_COLOR", "1")
	}

	// Create validator
	v := validator.NewSchemaValidator()

	// Register custom types if specified
	if *registerTypes != "" {
		for _, typeName := range strings.Split(*registerTypes, ",") {
			typeName = strings.TrimSpace(typeName)
			if typeName != "" {
				v.RegisterType(typeName, validator.TypeDescriptor{
					Name:        typeName,
					Description: "Custom type",
				})
			}
		}
	}

	// Process each file
	exitCode := 0
	for _, filename := range flag.Args() {
		if err := validateFile(filename, v); err != nil {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

func validateFile(filename string, v *validator.SchemaValidator) error {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		return err
	}

	sourceText := string(content)

	// Parse schema
	var schemaNode ast.SchemaNode
	var format parser.Format

	if *formatFlag == "auto" {
		// Auto-detect format from file extension or content
		format = detectFormatFromFile(filename, sourceText)
	} else {
		format = parseFormatString(*formatFlag)
	}

	if *verboseFlag {
		fmt.Fprintf(os.Stderr, "Parsing %s as %s format...\n", filename, format)
	}

	schemaNode, err = shape.Parse(format, sourceText)
	if err != nil {
		if *outputFlag != "quiet" {
			fmt.Fprintf(os.Stderr, "Parse error in %s: %v\n", filename, err)
		}
		return err
	}

	if *verboseFlag {
		fmt.Fprintf(os.Stderr, "Validating %s...\n", filename)
	}

	// Validate
	result := v.ValidateAll(schemaNode, sourceText)

	// Output results
	switch *outputFlag {
	case "json":
		jsonBytes, _ := result.ToJSON()
		if flag.NArg() > 1 {
			fmt.Printf("File: %s\n", filename)
		}
		fmt.Println(string(jsonBytes))
	case "quiet":
		// No output, just exit code
	default: // "text"
		if result.Valid {
			if *verboseFlag || flag.NArg() > 1 {
				fmt.Printf("%s: Valid\n", filename)
			} else {
				fmt.Println(result.FormatColored())
			}
		} else {
			if flag.NArg() > 1 {
				fmt.Printf("%s:\n", filename)
			}
			fmt.Println(result.FormatColored())
			if *verboseFlag {
				fmt.Fprintf(os.Stderr, "\nValidation failed with %d error(s)\n", result.ErrorCount())
			}
		}
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func detectFormatFromFile(filename string, content string) parser.Format {
	// First try file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jsonv":
		return parser.FormatJSONV
	case ".xmlv":
		return parser.FormatXMLV
	case ".yamlv", ".ymlv":
		return parser.FormatYAMLV
	case ".csvv":
		return parser.FormatCSVV
	case ".propsv":
		return parser.FormatPropsV
	case ".textv":
		return parser.FormatTEXTV
	}

	// Fall back to content detection
	format, err := parser.DetectFormat(content)
	if err != nil {
		// Default to JSONV if detection fails
		return parser.FormatJSONV
	}
	return format
}

func parseFormatString(formatStr string) parser.Format {
	switch strings.ToLower(formatStr) {
	case "jsonv":
		return parser.FormatJSONV
	case "xmlv":
		return parser.FormatXMLV
	case "yamlv":
		return parser.FormatYAMLV
	case "csvv":
		return parser.FormatCSVV
	case "propsv":
		return parser.FormatPropsV
	case "textv":
		return parser.FormatTEXTV
	default:
		fmt.Fprintf(os.Stderr, "Warning: unknown format '%s', using auto-detection\n", formatStr)
		return parser.FormatJSONV
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: shape-validate [flags] <schema-file>...

Validate Shape schema files for semantic correctness.

Flags:
  -f, --format string       Schema format (jsonv, xmlv, yamlv, csvv, propsv, textv, auto) (default "auto")
  -o, --output string       Output format (text, json, quiet) (default "text")
  --no-color                Disable colored output
  --register-type string    Register custom types (comma-separated)
  -v, --verbose             Verbose output
  --version                 Show version

Examples:
  # Validate a single file
  shape-validate schema.jsonv

  # Validate multiple files
  shape-validate schema1.jsonv schema2.xmlv

  # JSON output
  shape-validate -o json schema.jsonv

  # Register custom types
  shape-validate --register-type SSN,CreditCard schema.jsonv

  # Quiet mode (exit code only)
  shape-validate -o quiet schema.jsonv && echo "Valid!"

Exit Codes:
  0  Schema is valid
  1  Schema has validation errors
  2  Parse error (syntax error)
  3  File not found or I/O error
`)
}
