package configs

import (
	"io/ioutil"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestReadYamlConfig(t *testing.T) {
	yamlInput, err := ioutil.ReadFile("config.yml")
	if err != nil {
		t.Fatal(err)
	}

	conf := new(Configurations)
	if err := yaml.Unmarshal(yamlInput, &conf); err != nil {
		t.Fatal(err)
	}
	t.Logf("Read configs: %+v", conf)
}
