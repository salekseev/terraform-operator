
package terraform

import (
	"fmt"
	"bytes"
	"os"
	"os/exec"
	"io/ioutil"
	"encoding/json"
)

var TFPATH = os.Getenv("TFPATH")

type Provider struct {
	Provider map[string]interface{} `json:"provider"`
}

type Module struct {
	Module map[string]interface{} `json:"module"`
}

// RenderProviderToTerraform takes an object, and attempts to construct the appropriate terraform json from it.
func RenderProviderToTerraform(instance interface{}, providerName string) ([]byte, error) {
	r := Provider{
		Provider: map[string]interface{}{
			providerName: instance,
		},
	}
	b, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return b, err
	}
	return b, nil
}

// RenderModuleToTerraform takes an object, and attempts to construct the appropriate terraform json from it.
func RenderModuleToTerraform(instance interface{}, moduleName string) ([]byte, error) {
	r := Module{
		Module: map[string]interface{}{
			moduleName: instance,
		},
	}
	b, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return b, err
	}
	return b, nil
}

func WriteToFile(b []byte, name string) error {
	err := ioutil.WriteFile(TFPATH+"/"+name+".tf.json", b, 0755)
	if err != nil {
		return err
	}
	return nil
}

func TerraformInit() error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("terraform","init")
	cmd.Dir = TFPATH
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
    fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
    return err
	}

	fmt.Println("terraform init output:\n" + out.String())
	return nil
}

func TerraformValidate() error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("terraform","validate")
	cmd.Dir = TFPATH
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
    fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
    return err
	}

	fmt.Println("terraform validate output:\n" + out.String())
	return nil
}

func TerraformPlan() error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("terraform","plan")
	cmd.Dir = TFPATH
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
    fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
    return err
	}

	fmt.Println("terraform plan output:\n" + out.String())
	return nil
}

func TerraformApply() error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("terraform","apply","-auto-approve")
	cmd.Dir = TFPATH
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
    fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
    return err
	}

	fmt.Println("terraform apply output:\n" + out.String())
	return nil
}
