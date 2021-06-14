package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/cr-norton/tfconvert/pkg/aws"
	"github.com/cr-norton/tfconvert/pkg/codegen"
	"github.com/cr-norton/tfconvert/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	directory string
	save      bool
)

func main() {
	options, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	if options.StackName == "" {
		log.Fatal("stack name is required")
	}

	client, err := aws.New(options.Region)
	if err != nil {
		log.Fatalf("unable to configure aws client: %v", err)
	}

	ctx := context.Background()
	stack, err := client.GetStack(ctx, *options)
	if err != nil {
		log.Fatalf("unable to load cloudformation stack: %v", err)
	}

	save = true
	if save {
		bytes, _ := json.Marshal(stack)
		ioutil.WriteFile("stack.json", bytes, os.ModePerm)
	}

	tfout, err := codegen.Generate(stack, *options, aws.TemplateFunctions)
	if err != nil {
		log.Fatalf("unable to generate tf templates: %v", err)
	}

	err = writeFiles(directory, tfout)
	if err != nil {
		log.Fatal(err)
	}

	err = writeScripts(directory, *stack)
	if err != nil {
		log.Fatal(err)
	}

	err = terraFormat(directory)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("terraform migration complete")
}

func parseFlags() (*types.Options, error) {
	var config, stack, region, service string

	flag.StringVar(&config, "config", "", "config file location")
	flag.StringVar(&stack, "stack", "", "stack name")
	flag.StringVar(&region, "region", "", "aws region")
	flag.StringVar(&service, "service", "", "service name")
	flag.StringVar(&directory, "output", "./terraform", "output directory")
	flag.Parse()

	if config != "" {
		return loadConfig(config)
	}

	if stack == "" {
		return nil, errors.New("stack is required")
	}

	options := &types.Options{
		StackName:      stack,
		ServiceName:    service,
		Region:         region,
		AdditionalTags: map[string]string{},
	}
	if options.ServiceName == "" {
		options.ServiceName = options.StackName
	}
	return options, nil
}

func loadConfig(filename string) (*types.Options, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config file")
	}

	var options types.Options
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, errors.Wrap(err, "unable to parse codegen config")
	}
	return &options, nil
}

func writeFiles(directory string, files map[string]string) error {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return errors.Wrap(err, "unable to create output directory")
	}

	for file, content := range files {
		if len(content) == 0 {
			continue
		}
		file = fmt.Sprintf("%s/%s", directory, file)
		if err := ioutil.WriteFile(file, []byte(content), os.ModePerm); err != nil {
			return errors.Wrap(err, "unable to write file")
		}
	}

	return nil
}

func writeScripts(directory string, stack aws.Stack) error {
	commands, err := codegen.GenerateImportScript(stack.Resources())
	if err != nil {
		return err
	}

	file := fmt.Sprintf("%s/%s", directory, "import.sh")
	content := strings.Join(commands, "\n")
	if err := ioutil.WriteFile(file, []byte(content), os.ModePerm); err != nil {
		return errors.Wrap(err, "unable to write import script")
	}
	return nil
}

func terraFormat(directory string) error {
	if err := os.Chdir(directory); err != nil {
		return errors.Wrap(err, "unable to cd to output directory")
	}
	if err := exec.Command("terraform", "fmt").Run(); err != nil {
		return errors.Wrap(err, "unable to format terraform code")
	}
	return nil
}
