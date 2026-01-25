package llm

import (
	"Hyper/pkg/log"
	"context"
	"fmt"
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
func GenNoteTag(ctx context.Context, ossURL string) []string {

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
	startTime := time.Now()
	userMessage := openai.ChatCompletionUserMessageParam{
		Content: openai.ChatCompletionUserMessageParamContentUnion{
			OfArrayOfContentParts: contentParts,
		},
	}
	params := openai.ChatCompletionNewParams{
		Model: "qwen3-vl-plus",
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfUser: &userMessage},
		},
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

func ClassifyMultiImageNote(ctx context.Context, title, content string, ossURLs []string) string {
	channels := []string{
		"滑板", "骑行", "派对", "纹身", "改装车", "露营",
		"篮球", "足球", "飞盘", "潮鞋", "电子竞技", "健身", "艺术",
	}
	channelList := strings.Join(channels, "、")
	promptText := fmt.Sprintf(
		"你是一个内容分类专家。请结合提供的文字和图片，从以下列表中选择一个最贴切的频道返回。\n\n"+
			"【标题】：%s\n"+
			"【正文】：%s\n\n"+
			"候选频道列表：%s\n"+
			"要求：直接输出频道名称，不要解释，不要带标点。",
		title, content, channelList,
	)
	contentParts := []openai.ChatCompletionContentPartUnionParam{
		{
			OfText: &openai.ChatCompletionContentPartTextParam{
				Text: promptText,
			},
		},
	}
	for _, url := range ossURLs {
		url = url + "?x-oss-process=image/resize,w_200"
		contentParts = append(contentParts, openai.ChatCompletionContentPartUnionParam{
			OfImageURL: &openai.ChatCompletionContentPartImageParam{
				ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
					URL: url,
				},
			},
		})
	}
	contentParts = append(contentParts, openai.ChatCompletionContentPartUnionParam{
		OfText: &openai.ChatCompletionContentPartTextParam{},
	})
	startTime := time.Now()
	userMessage := openai.ChatCompletionUserMessageParam{
		Content: openai.ChatCompletionUserMessageParamContentUnion{
			OfArrayOfContentParts: contentParts,
		},
	}
	params := openai.ChatCompletionNewParams{
		Model: "qwen3-vl-plus",
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfUser: &userMessage},
		},
	}
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.L.Error("failed to gen tag", zap.Error(err))
		return ""
	}
	Content := completion.Choices[0].Message.Content
	log.L.Info("gen tag", zap.String("tag", Content), zap.Duration("gen time", time.Since(startTime)))
	return Content
}
