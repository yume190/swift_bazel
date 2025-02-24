package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/cgrindel/swift_bazel/tools/generate_ci_workflow/internal/github"
	"github.com/cgrindel/swift_bazel/tools/generate_ci_workflow/internal/testparams"
	"gopkg.in/yaml.v3"
)

const intTestMatrixKey = "integration_test_matrix"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stderr *os.File) error {
	var (
		templatePath          string
		intTestParamsJSONPath string
		outputPath            string
	)
	flag.StringVar(&templatePath, "template", "", "path to the template file")
	flag.StringVar(&intTestParamsJSONPath, "int_test_params_json", "", "path to the test params JSON file")
	flag.StringVar(&outputPath, "output", "", "path to the output file")
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), `usage: bazel run //tools/generate_ci_workflow -- -template <template_path> -int_test_params_json <int_test_params_json> -output <output>

This utility generates a new GitHub actions workflow file for this project.

`)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Read the workflow YAML
	workflowYAML, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template at '%s': %w", templatePath, err)
	}
	workflow, err := github.NewWorkflowFromYAML(workflowYAML)
	if err != nil {
		return err
	}

	// Read the example JSON
	intTestParamsJSON, err := os.ReadFile(intTestParamsJSONPath)
	if err != nil {
		return fmt.Errorf(
			"could not read integration test params JSON at '%s': %w",
			intTestParamsJSONPath,
			err,
		)
	}
	intTestParams, err := testparams.NewIntTestParamsFromJSON(intTestParamsJSON)
	if err != nil {
		return err
	}

	// Set up the macOS matrix
	if err := updateJob(workflow.Jobs, intTestMatrixKey, intTestParams); err != nil {
		return err
	}

	// Write the output file
	var outBuf bytes.Buffer
	if _, err := outBuf.WriteString(hdrMsg); err != nil {
		return err
	}
	yamlEncoder := yaml.NewEncoder(&outBuf)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&workflow)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outBuf.Bytes(), 0666); err != nil {
		return fmt.Errorf("failed to write output YAML to '%s': %w", outputPath, err)
	}

	return nil
}

func updateJob(jobs map[string]github.Job, key string, intTestParams []testparams.IntTestParams) error {
	job, ok := jobs[key]
	if !ok {
		return fmt.Errorf("did not find '%s' job", key)
	}
	matrix := job.Strategy.Matrix
	updateMatrix(&matrix, intTestParams)
	job.Strategy.Matrix = matrix
	jobs[key] = job

	return nil
}

func updateMatrix(m *github.SBMatrixStrategy, intTestParams []testparams.IntTestParams) {
	newM := github.SBMatrixStrategy{}
	for _, itp := range intTestParams {
		inc := github.SBMatrixInclude{
			Test:         itp.Test,
			Runner:       itp.Runner(),
			EnableBzlmod: itp.EnableBzlmod(),
		}
		newM.Include = append(newM.Include, inc)
	}
	*m = newM
}

const hdrMsg = `# Portions of this file are generated by the build.  
#
# Note:
# - Modification to values outside of the matrix strategy sections should 
#   persist.
# - Comments and custom formatting will be lost.
`
