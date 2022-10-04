package runner

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func runProcessor(cfg ProcessorConfig, filePath, codePath string) (*Result, error) {
	if cfg.SkipProcessing {
		log.Println("Skipping processing of the test script output")

		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return unmarshalResult(b, codePath)
	}

	log.Printf("Processing the test script output using %q\n", cfg.Interpreter)

	cmd := exec.Command(cfg.Interpreter)
	cmd.Env = append(os.Environ(), "INPUT_FILE="+filePath)

	buf := &bytes.Buffer{}
	cmd.Stdin = strings.NewReader(cfg.Script)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return unmarshalResult(buf.Bytes(), codePath)
}
