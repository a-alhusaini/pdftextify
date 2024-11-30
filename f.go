package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	openai "github.com/openai/openai-go" // imported as openai
)

type Transcript struct {
	Text string `json:"text" jsonschema_description:"Transcription of the provided image"`
}

var TranscriptSchema = GenerateSchema[Transcript]()

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("no OPENAI_API_KEY in environment")
		os.Exit(1)
	}

	useNew := flag.Bool("n", false, "use new version")
	flag.Parse()
	if *useNew {
		newVersion()
	} else {
		OldVersion()
	}
}

func newVersion() {
	document := flag.Args()[len(flag.Args())-1]
	outputPath := filepath.Join("outputs", document+"_data")

	clearOutputDir(outputPath)
}

func clearOutputDir(outputPath string) {
	err := os.RemoveAll(outputPath)
	if err != nil && err != os.ErrNotExist {
		panic(err)
	}

	os.MkdirAll(outputPath, 0700)
}

func OldVersion() {
	fname := os.Args[1]

	b64File := fileToB64(fname)

	fmt.Println(GetTranscript(b64File))
}

func GetTranscript(b64File string) Transcript {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("Transcript"),
		Description: openai.F("Transcript of an image"),
		Schema:      openai.F(TranscriptSchema),
		Strict:      openai.Bool(true),
	}

	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("transcribe the following image:"),
			openai.UserMessageParts(openai.ImagePart("data:image/jpeg;base64," + string(b64File))),
		}),
		MaxCompletionTokens: openai.F[int64](16384),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		panic(err.Error())
	}

	r := Transcript{}

	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &r)
	if err != nil {
		panic(err)
	}

	return r
}

func fileToB64(fname string) string {
	fileData, err := os.ReadFile(fname)
	if err != nil {
		out := fmt.Sprintf("failed to read file %s\n", err)
		panic(out)
	}

	b64File := base64.StdEncoding.EncodeToString(fileData)
	return b64File
}
