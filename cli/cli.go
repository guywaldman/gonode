package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/guywaldman/gonode/generator"
	"github.com/guywaldman/gonode/internal/config"
	"github.com/guywaldman/gonode/parser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

type Args struct {
	Dir    *string
	Output *string
}

func main() {
	//exhaustruct:ignore
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	//exhaustruct:ignore
	app := &cli.App{
		Name:  "gonode",
		Usage: "CLI tool for building Node.js addons using Go",
		Flags: []cli.Flag{
			//exhaustruct:ignore
			&cli.StringFlag{
				Name:  "dir",
				Usage: "The directory to parse",
			},
			//exhaustruct:ignore
			&cli.StringFlag{
				Name:  "output",
				Usage: "The output directory",
			},
		},
		Action: func(c *cli.Context) error {
			return runApp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Error running app")
	}
}

func runApp(ctx *cli.Context) error {
	dir := ctx.String("dir")
	output := ctx.String("output")

	log.Info().Msgf("Received arguments: dir=%v, output=%v", dir, output)

	configPath := filepath.Join(dir, ".gonode.yaml")
	reader, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error opening config:", err)
		os.Exit(1)
	}

	config, err := config.NewFromReader(reader)
	if err != nil {
		fmt.Println("Error parsing config:", err)
		os.Exit(1)
	}

	var filesToParse []string
	for _, file := range config.Files {
		filesFromGlob, err := filepath.Glob(path.Join(dir, file))
		if err != nil {
			fmt.Println("Error compiling glob:", err)
			os.Exit(1)
		}

		filesToParse = append(filesToParse, filesFromGlob...)
	}

	log.Info().Msgf("Parsing files: %v", strings.Join(filesToParse, ", "))

	var exportedFunctions []parser.ExportedFunction
	for _, file := range filesToParse {
		parser, err := parser.New()
		if err != nil {
			fmt.Println("Error creating parser:", err)
			os.Exit(1)
		}
		reader, err := os.Open(file)
		if err != nil {
			fmt.Println("Error opening file:", err)
			os.Exit(1)
		}
		exported, err := parser.Parse(file, reader)
		if err != nil {
			log.Error().Err(err).Msgf("Error parsing file: %v", file)
			os.Exit(1)
		}
		for _, exportedFunction := range exported {
			exportedFunctions = append(exportedFunctions, *exportedFunction)
		}
	}

	if len(exportedFunctions) == 0 {
		log.Info().Msg("No exported functions found")
		return nil
	}

	log.Info().Msg("Generating bindings for exported functions:")
	generator, err := generator.New(config.AddonName, config.OutputDir)
	if err != nil {
		log.Error().Err(err).Msg("Error creating generator")
		os.Exit(1)
	}
	generated, err := generator.Generate(exportedFunctions)
	if err != nil {
		log.Error().Err(err).Msg("Error generating bindings")
		os.Exit(1)
	}

	// Normlize the output directory
	outputDir := path.Join(dir, config.OutputDir, "gonode")
	cppOutputPath := path.Join(outputDir, config.AddonName+".cc")
	log.Info().Msgf("Writing Node.js addon C++ code to %v", cppOutputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Error().Err(err).Msg("Error creating output directory")
		os.Exit(1)
	}

	// Empty the output directory
	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Error().Err(err).Msg("Error reading output directory")
		os.Exit(1)
	}
	for _, file := range files {
		err = os.Remove(path.Join(outputDir, file.Name()))
		if err != nil {
			log.Error().Err(err).Msg("Error removing file")
			os.Exit(1)
		}
	}

	// Write the C++ code
	err = os.WriteFile(cppOutputPath, []byte(generated.CppCode), 0644)
	if err != nil {
		log.Error().Err(err).Msg("Error writing C++ code")
		os.Exit(1)
	}
	log.Info().Msg("Successfully generated Node.js addon")

	// Write the TypeScript file
	typeScriptDefinitionsPath := path.Join(outputDir, config.AddonName+".ts")
	log.Info().Msgf("Writing TypeScript definitions to %v", typeScriptDefinitionsPath)
	err = os.WriteFile(typeScriptDefinitionsPath, []byte(generated.TypeScriptCode), 0644)
	if err != nil {
		log.Error().Err(err).Msg("Error writing TypeScript definitions")
		os.Exit(1)
	}
	log.Info().Msg("Successfully generated TypeScript definitions")

	return nil
}
