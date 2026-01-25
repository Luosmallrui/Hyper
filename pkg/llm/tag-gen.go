package llm

import (
	"Hyper/pkg/log"
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"go.uber.org/zap"
)

var client openai.Client

func init() {
	client = openai.NewClient(
		option.WithAPIKey("sk-798f3a22651446b1b4c441675dea02eb"),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	)
}
func GenNoteTag(ossURL string) []string {

	// 初始化 Client
	contentParts := []openai.ChatCompletionContentPartUnionParam{
		{
			// 第一个元素：纯文本
			OfText: &openai.ChatCompletionContentPartTextParam{
				Text: "作为小红书专家，只输出5个小红书话题标签，用#开头，用空格分隔，不要任何其他内容",
			},
		},
		{
			// 第二个元素：纯图片
			OfImageURL: &openai.ChatCompletionContentPartImageParam{
				ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
					URL: ossURL,
				},
			},
		},
	}
	ctx := context.Background()
	startTime := time.Now()
	userMessage := openai.ChatCompletionUserMessageParam{
		Content: openai.ChatCompletionUserMessageParamContentUnion{
			OfArrayOfContentParts: contentParts,
		},
	}
	params := openai.ChatCompletionNewParams{
		Model: "qwen3-vl-flash",
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfUser: &userMessage},
		},
		MaxTokens: openai.Int(50),
	}
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.L.Error("failed to gen tag", zap.Error(err))
		return make([]string, 0)
	}
	Content := completion.Choices[0].Message.Content
	log.L.Info("gen tag", zap.String("tag", Content), zap.Duration("gen time", time.Since(startTime)))
	return ParseTags(Content)
}

func ParseTags(input string) []string {
	re := regexp.MustCompile(`#[^\s#]+`)
	matches := re.FindAllString(input, -1)

	var tags []string
	for _, tag := range matches {
		cleanTag := strings.TrimPrefix(tag, "#")
		tags = append(tags, cleanTag)
	}
	return tags
}
