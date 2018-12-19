package readwrite

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	taask "github.com/taask/client-golang"
	yaml "gopkg.in/yaml.v2"
)

// ReadTaskFile reads a file and converts it to a task
func ReadTaskFile(filepath string) (*taask.Task, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadFile")
	}

	spec := &taask.Spec{}
	if err := yaml.Unmarshal(raw, spec); err != nil {
		if jsonErr := json.Unmarshal(raw, spec); jsonErr != nil {
			return nil, errors.Wrap(jsonErr, errors.Wrap(err, "failed to yaml and json Unmarshal").Error()) // stupid, but whatever
		}
	}

	return &spec.Spec, nil
}
