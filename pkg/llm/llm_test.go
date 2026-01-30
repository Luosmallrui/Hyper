package llm

import (
	"context"
	"testing"
)

func TestClassifyMultiImageNote(t *testing.T) {
	// 1. 准备测试数据
	// 注意：确保这些 URL 是公网可访问的，或者替换为你 OSS 里的真实测试图片
	testTitle := "周末去龙泉山，风景绝了！"
	testContent := "今天天气不错，和朋友一起骑行了50公里，虽然累但很开心。#骑行 #户外"
	testImages := []string{
		"https://cdn.hypercn.cn/note/2026/01/25/2015344440675143680.jpg",
		"https://cdn.hypercn.cn/note/2026/01/25/2015344440926801920.jpg",
		"https://cdn.hypercn.cn/note/2026/01/25/2015344441170071552.jpg",
	}

	// 3. 执行测试
	ctx := context.Background()
	channel := ClassifyMultiImageNote(ctx, testTitle, testContent, testImages, testImages)

	// 4. 验证结果
	if channel == "" {
		t.Error("分类结果为空，请检查网络或 API 调用")
	}

	// 验证结果是否在你的白名单内
	validChannels := map[string]bool{
		"骑行": true, "滑板": true, "改装车": true, // ... 其他频道
	}

	if !validChannels[channel] {
		t.Errorf("预测结果 [%s] 不在预设频道列表中", channel)
	}

	t.Logf("测试成功！识别出的频道为: %s", channel)
}
