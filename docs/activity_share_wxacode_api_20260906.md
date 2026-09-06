# 活动分享海报小程序码接口

更新时间：2026-09-06

Base URL：`/api/v1`

本文面向后端，说明小程序「活动购票分享海报」所需的小程序码生成接口。海报效果：上半部分为活动海报图，下方为活动标题/日期/城市，右下角放置圆形小程序码，观众微信扫码后直达活动详情页购票。

## 1. 背景与约束

- 海报上的码必须是**微信小程序码**（圆形菊花码），扫码后直接打开小程序指定页面。
- 该码只能通过微信服务端接口 [`wxacode.getUnlimited`](https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/wxacode.getUnlimited.html) 生成，调用需要 `access_token`，**小程序前端无法自行生成**，因此需要后端提供接口。
- 普通二维码（编码 URL 文本）微信扫码后只能打开网页，不能打开小程序，无法复用订单入场码的生成方式。

## 2. 接口定义

```http
GET /api/v1/activity/:id/wxacode
Authorization: Bearer <access_token>（建议允许游客访问，与活动详情可见性保持一致）
```

响应（复用现有 CDN 返回 URL 的模式，与 `/upload` 一致）：

```json
{
  "code": 200,
  "data": {
    "url": "https://cdn.hypercn.cn/ticketing/wxacode/activity/86.png"
  }
}
```

要求：

- **必须转存 CDN 返回 URL**，不要直接返回微信接口的图片二进制流。前端需要用 `<Image>` 直接显示、并绘制到 canvas 上，CDN URL 无跨域和域名白名单问题。
- **按活动缓存**：同一活动的小程序码内容不变，生成一次后复用即可。可按 `activity_id` 命名文件（重复请求直接返回已有 URL），或在活动表加 `wxacode_url` 字段。活动上下架不影响码的有效性。
- 活动不存在时按既有规范返回 `code: 404`。

## 3. 微信侧调用参数

| 参数 | 值 | 说明 |
|---|---|---|
| `page` | `pages/activity/index` | 活动详情页路径，已在小程序已发布版本中存在。 |
| `scene` | 活动 ID，如 `86` | 纯数字，远低于 32 字符上限。扫码进入时前端从 `query.scene` 读取活动 ID。 |
| `check_path` | `false` | 避免页面未发布时生成失败，体验版/开发版也能正常生成。 |
| `env_version` | `release` | 线上固定用正式版；联调阶段可临时用 `trial`（体验版）验证，上线前切回。 |
| `width` | `430` | 可选项，建议 430（默认）或 860（高清）。海报上展示尺寸约 110pt，430 已够清晰。 |
| `auto_color` / `line_color` | 默认黑色即可 | 海报为深色底时前端会加白色衬底，不需要改码的颜色。 |

## 4. 前端配合点（后端无需处理，仅同步信息）

- 扫码进入路径：`pages/activity/index?scene=86`，前端会在活动详情页 `onLoad` 中优先解析 `query.scene` 作为活动 ID，兼容现有 `?id=86` 分享链接。
- 海报图片由前端在小程序内用 canvas 合成（海报图 + 活动信息 + 小程序码），用户保存到相册后自行宣发，后端无需参与海报生成。

## 5. 可选扩展（本期不做，预留说明）

如需统计「谁分享的海报带来了扫码/订单」， `scene` 可扩展为 `活动ID_分享者ID` 形式（如 `86_12345`），总长仍远低于 32 字符限制。前端解析时按下划线拆分第一段即可，旧格式的码保持兼容。本期 `scene` 只传活动 ID。

## 6. 验收标准

1. 调用接口返回 `code: 200` 且 `url` 为可公开访问的 CDN 图片地址。
2. 微信「扫一扫」扫描该码，直接进入小程序正式版活动详情页并展示对应活动。
3. 同一活动重复调用返回相同 URL，不重复消耗微信接口配额。

## 7. 后端实现说明（2026-09-06 已上线）

- 路由：`GET /api/v1/activity/:id/wxacode`，挂 `optionalAuth`（游客可访问，与活动详情一致）。
- 实现：`service/ticketing.go` `GetActivityWxacode`——校验活动存在（不存在返回 404）→ OSS `HeadObject` 判断 `ticketing/wxacode/activity/<id>.png` 是否已存在（存在直接返回 URL，不消耗微信配额）→ 否则调 `wxacode.getUnlimited`（`scene=活动ID`、`page=pages/activity/index`、`check_path=false`、`width=430`）→ `UploadRaw` 转存 OSS → 返回 CDN URL。
- `env_version` 走配置 `app.qr_code_env_version`（缺省 `trial`，正式上线后配置为 `release`）。
- 配套改动：`IOssService` 新增 `ObjectExists` / `CDNUrl` 两个方法；清理了 `GenerateUnlimitedQRCode` 里打印含 access_token URL 的调试输出。
