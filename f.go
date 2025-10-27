package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	// imported as openai
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
	apiKey := os.Getenv("GROQ_API_KEY")
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

	ts := GetTranscript(b64File)
	fmt.Println(ts)
}

func GetTranscript(b64File string) Transcript {
	ctx := context.Background()

	data := []byte(fmt.Sprintf(`{
  "messages": [
    {
      "role": "user",
      "content": [
	 	{
	  		"type": "text",
			"text": "Transcribe the following image"
		},
		{
			"type": "image_url",
			"image_url": {
				"url": "data:image/jpeg;base64,%s"
			}
		}
	  ]
    }
  ],
  "model": "meta-llama/llama-4-maverick-17b-128e-instruct",
  "temperature": 1,
  "max_completion_tokens": 4096,
  "top_p": 1,
  "stream": false,
  "response_format": {
    "type": "json_schema",
	"json_schema": {
		"name": "transcript",
		"schema": {
			"type": "object",
			"properties": {
				"text": {"type": "string"}
			}
		}
	}
  },
  "stop": null
}
	`, b64File))

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Groq API request failed: %v", err))
	}
	defer res.Body.Close()

	out, _ := io.ReadAll((res.Body))

	prettyJSON := map[string]any{}

	json.Unmarshal(out, &prettyJSON)
	out, _ = json.MarshalIndent(prettyJSON, "", "    ")

	choices := prettyJSON["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	t := Transcript{}
	_ = json.Unmarshal([]byte(content), &t)
	return t
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
