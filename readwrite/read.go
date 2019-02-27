package readwrite

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	taask "github.com/taask/client-golang"
	yaml "gopkg.in/yaml.v2"
)

// ReadTaskSpecFile reads a file and converts it to a task
func ReadTaskSpecFile(filepath string) (*taask.Task, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadFile")
	}

	spec := taask.Spec{}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		if jsonErr := json.Unmarshal(raw, &spec); jsonErr != nil {
			return nil, errors.Wrap(jsonErr, errors.Wrap(err, "failed to yaml and json Unmarshal").Error()) // stupid, but whatever
		}
	}

	return &spec.Spec, nil
}

// ReadTaskSpecFromStdin returns a task piped in
func ReadTaskSpecFromStdin() (*taask.Task, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "failed to Stdin.Stat")
	}

	raw := []byte{}

	if (stat.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe {
		r := bufio.NewReader(os.Stdin)

		for {
			line, _, err := r.ReadLine()
			if err != nil {
				break
			}

			raw = append(raw, line...)
			raw = append(raw, []byte("\n")...)
		}
	}

	spec := &taask.Spec{}
	if err := yaml.Unmarshal(raw, spec); err != nil {
		if jsonErr := json.Unmarshal(raw, spec); jsonErr != nil {
			return nil, errors.Wrap(jsonErr, errors.Wrap(err, "failed to yaml and json Unmarshal").Error()) // stupid, but whatever
		}
	}

	return &spec.Spec, nil
}
