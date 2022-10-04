package runner

import (
	"os"
)

// setEnv sets the environment variables using the provided key-value map
func setEnv(env map[string]string) error {
	for k, v := range env {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// unsetEnv unsets the environment variables using keys from the provided
// key-value map
func unsetEnv(env map[string]string) error {
	for k := range env {
		err := os.Unsetenv(k)
		if err != nil {
			return err
		}
	}

	return nil
}
