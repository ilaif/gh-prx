package ai

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"

	"github.com/ilaif/gh-prx/pkg/config"
)

func IsAISummarizerAvailable() bool {
	return config.GetOpenAIApiKey() != ""
}

func SummarizeGitDiffOutput(ctx context.Context, diffOutput string) (string, error) {
	client := openai.NewClient(config.GetOpenAIApiKey())

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
					Role: openai.ChatMessageRoleUser,
					Content: heredoc.Docf(`Please summarize the pull request.
						Write your response in numbered points.
						Write a high level description.
						Do not repeat the commit summaries or the file summaries.
						Write the most important bullet points.
						The list should not be more than a few bullet points.
						Mention the file names that were changed, if applicable.
						The git diff changes for the PR:
						%s
					`, diffOutput),
				},
			},
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "Failed to summarize git diff output using AI")
	}

	return resp.Choices[0].Message.Content, nil
}
