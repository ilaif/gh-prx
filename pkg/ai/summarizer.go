package ai

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"

	"github.com/ilaif/gh-prx/pkg/config"
)

func IsAISummarizerAvailable() bool {
	return config.GetOpenAIApiKey() != ""
}

func SummarizeGitDiffOutput(ctx context.Context, diffOutput string, prBody string) (string, error) {
	client := openai.NewClient(config.GetOpenAIApiKey())

	userPrompt := heredoc.Docf(`Please summarize the pull request changes.

		The git diff output for the PR:
		'''
		%s
		'''

		Structure your answer to conform with the following template:
		'''
		%s
		'''

		Please follow these guidelines:
		- Do not repeat the commit summaries or the file summaries.
		- Mention the file names that were changed, if applicable.
		- Prefer bullet points over long sentences.
	`, diffOutput, prBody)

	log.Debug(fmt.Sprintf("Creating an AI-powered summary based on prompt:\n%s", userPrompt))

	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 1024,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are a code reviewer. You are summarizing a pull request according to the code changes. " +
						"You like making descriptions short and to the point.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "Failed to summarize git diff output using AI")
	}

	return resp.Choices[0].Message.Content, nil
}
