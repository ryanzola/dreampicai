package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/replicate/replicate-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// You can also provide a token directly with
	// `replicate.NewClient(replicate.WithToken("r8_..."))`
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	// https://replicate.com/stability-ai/stable-diffusion
	// version := "stability-ai/stable-diffusion:ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4"

	input := replicate.PredictionInput{
		"prompt": "multicolor hyperspace",
	}

	webhook := replicate.Webhook{
		URL:    "https://webhook.site/6153d744-bf14-4139-8d21-f008e91e8a74",
		Events: []replicate.WebhookEventType{"start", "completed"},
	}

	// Run a model and wait for its output
	output, err := r8.CreatePrediction(ctx, "ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4", input, &webhook, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output: ", output)
}
