# 微信小程序内容安全接入说明

更新时间：2026-07-24

## 目标

所有小程序用户可发布的文本在写入数据库前调用微信 `msgSecCheck` 文本内容安全接口。命中风险内容不写库，客户端只收到：

```text
内容含违规信息
```

不返回命中标签、关键词、微信错误码或审核细节。

## 已覆盖的发布场景

| 场景 | 接口 | 微信 scene | 校验字段 |
|---|---|---:|---|
| 评论和回复 | `POST /api/v1/comments/create` | `2` | `content` |
| 用户发布动态 | `POST /api/v1/note/create` | `4` | `title` + `content` |
| 商家发布动态 | `POST /api/v1/organizer/posts` | `4` | `title` + `content` |
| 商家编辑动态 | `PUT /api/v1/organizer/posts/:id` | `4` | `title` + `content` |
| 用户昵称、签名 | `POST /api/v1/auth/update-profile` | `1` | `username`、`motto`、`signature` |

## 后端调用

调用地址：

```http
POST https://api.weixin.qq.com/wxa/msg_sec_check?access_token={access_token}
Content-Type: application/json
```

请求使用 V2 参数：

```json
{
  "content": "用户发布的文本",
  "version": 2,
  "openid": "发布者的小程序openid",
  "scene": 2,
  "nickname": "用户昵称",
  "title": "可选标题"
}
```

处理规则：

- `errcode = 0` 且 `result.suggest = pass`：允许发布。
- `errcode = 87014`，或 `result.suggest` 不是 `pass`：拒绝发布，提示“内容含违规信息”。
- 微信接口不可用、缺少 `openid`、令牌获取失败或响应异常：拒绝发布，提示“内容安全验证失败，请稍后重试”。
- 文本按微信接口限制控制在 2500 字符以内；超长文本不写库。

## 部署前确认

现有配置必须有有效的小程序凭证：

```yaml
app:
  app_id: "wx..."
  app_secret: "..."
```

发布内容的用户必须通过小程序微信登录产生并保存 `users.open_id`。缺少 `openid` 的历史手机号账号不能绕过校验；应先完成微信登录后再发布。

微信官方接口参考：[文本内容安全](https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.msgSecCheck.html)。

## 重新提审前清理

本次代码只拦截新发布内容，不会自动删除数据库里已有的违规评论。审核人员留下的“澳门赌场在线发牌”评论需要先在管理端隐藏或删除：

```http
PATCH /api/v1/admin/note-comments/:comment_id/status
Authorization: Bearer {admin_access_token}
Content-Type: application/json

{
  "status": 0,
  "reason": "违规内容清理"
}
```

清理完成后，使用违规文本测试评论发布应立即失败并只提示“内容含违规信息”；使用正常文本应可正常发布。两项均验证通过后再提交小程序审核。
