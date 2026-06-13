# 核销员扫码绑定流程说明

更新时间：2026-06-11

本文档说明“后台添加核销员后，由核销员扫码进入小程序并绑定账号”的完整流程。

## 1. 业务目标

核销员不是由后台直接创建小程序账号。

正确流程是：

1. 主办方后台添加核销员姓名和手机号。
2. 后台生成核销员激活码。
3. 核销员扫码跳转到小程序。
4. 小程序判断当前用户是否已登录/注册。
5. 未注册用户先完成手机号注册/登录。
6. 小程序带登录态调用激活接口。
7. 后端把核销员邀请记录绑定到当前小程序用户。

绑定完成后，`verifiers.user_id` 会记录小程序用户 ID，`verifiers.bound_at` 会记录绑定时间。

## 2. 数据模型调整

`verifiers` 表新增字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| user_id | bigint unsigned | 已绑定的小程序用户 ID，未绑定为 0 |
| bound_at | datetime | 绑定完成时间 |

上线 DDL：

```sql
ALTER TABLE `verifiers` ADD COLUMN `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定的小程序用户ID' AFTER `organizer_id`;
ALTER TABLE `verifiers` ADD COLUMN `bound_at` datetime NULL COMMENT '绑定时间' AFTER `channel`;
ALTER TABLE `verifiers` ADD KEY `idx_verifier_user` (`user_id`) USING BTREE;
```

## 3. 主办方后台流程

### 3.1 添加核销员

```http
POST /api/v1/organizer/verifier
Authorization: Bearer <organizer_token>
```

请求：

```json
{
  "name": "核销员A",
  "phone": "13800138000"
}
```

说明：

- 后台只需要提交姓名和手机号。
- 新增后核销员状态为未激活。
- 此时不会创建小程序用户，也不会写入 `user_id`。

### 3.2 获取激活码

```http
GET /api/v1/organizer/verifier/:id/activation-qr
Authorization: Bearer <organizer_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "wechat_mini_program_code_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_page": "pages/user-sub/verifier-bind/index",
    "wechat_scene": "v=1",
    "wechat_qr": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "douyin_qr": "hyper://verifier/activate?verifier_id=1&channel=douyin"
  }
}
```

说明：

- `verifier_id` 是核销员邀请记录 ID。
- `wechat_qr` / `wechat_qr_url` / `wechat_mini_program_code_url` 都是后端调用微信小程序码接口生成并上传 OSS 后的图片地址，后台页面优先展示这些图片 URL。
- `wechat_scene` 是微信小程序码携带的短参数，小程序页面需要从 `scene` 里解析 `v`。
- 微信小程序码中不放手机号，避免手机号暴露。
- `douyin_qr` 仍是抖音侧占位深链，和微信小程序码无关。

## 4. 小程序绑定流程

小程序扫码进入后按以下流程处理：

1. 解析微信 `scene` 参数：`v`。
2. 判断当前用户是否已有登录态。
3. 如果未登录，先走现有手机号登录/注册流程。
4. 登录成功后，拿到 `access_token`。
5. 调用核销员激活接口完成绑定。

微信小程序码参数示例：

```text
scene=v%3D1
```

小程序解析后映射为：

```text
verifier_id = v
channel = wechat
```

微信小程序码不会直接跳转到带 query 的完整链接，例如：

```text
/pages/user-sub/verifier-bind/index?organizerName=测试主办方
```

实际打开的是 `pages/user-sub/verifier-bind/index`，参数在 `options.scene` 中。由于微信 `scene` 有长度限制，且中文主办方名称编码后很容易超长，二维码中只放短参数 `v`。页面需要通过 `v` 查询确认信息。

### 4.1 获取绑定确认信息

```http
GET /api/v1/verifier/activation-info?v=1
```

或：

```http
GET /api/v1/verifier/activation-info?verifier_id=1
```

响应：

```json
{
  "code": 200,
  "data": {
    "verifier_id": 1,
    "name": "核销员A",
    "phone": "13800138000",
    "status": 0,
    "is_bound": false,
    "organizer_id": 10,
    "organizer_name": "测试主办方"
  }
}
```

前端绑定页可用 `organizer_name` 展示确认文案。

### 4.2 激活并绑定核销员

```http
POST /api/v1/verifier/activate
Authorization: Bearer <access_token>
```

请求：

```json
{
  "verifier_id": 1,
  "phone": "13800138000",
  "channel": "wechat"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "verifier_id": 1,
    "user_id": 10001,
    "status": 1
  }
}
```

说明：

- 本接口必须带小程序用户登录态。
- 如果用户没有注册，需要先通过现有手机号登录/注册接口完成注册。
- 后端会校验登录用户的 `users.mobile` 是否等于请求里的 `phone`。
- 后端还会校验 `verifier_id` 对应的核销员手机号是否等于请求里的 `phone`。

## 5. 后端校验规则

激活接口会执行以下校验：

| 场景 | 结果 |
| --- | --- |
| 用户未登录 | 鉴权失败 |
| 登录用户没有绑定手机号 | 返回 `请先绑定手机号` |
| 登录用户手机号和核销员手机号不一致 | 返回 `登录手机号与核销员手机号不一致` |
| `verifier_id + phone` 找不到邀请记录 | 返回 `核销员邀请不存在` |
| 该邀请已经绑定其他用户 | 返回 `该核销员已绑定其他账号` |
| 同一用户重复扫码绑定同一邀请 | 幂等成功 |

## 6. 绑定成功后的数据变化

绑定成功后，`verifiers` 表会更新：

```text
user_id = 当前登录用户ID
status = 1
channel = 请求传入的 channel
bound_at = 当前时间
```

## 7. 微信小程序码生成方式

正式微信核销码由后端生成，不由后台前端用普通二维码库生成。

### 7.1 生成时机

主办方后台调用以下接口时，后端会实时生成小程序码：

```http
GET /api/v1/organizer/verifier/:id/activation-qr
Authorization: Bearer <organizer_token>
```

### 7.2 后端生成参数

后端会调用微信接口：

```http
POST https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=<access_token>
```

请求体：

```json
{
  "scene": "v=1",
  "page": "pages/user-sub/verifier-bind/index",
  "check_path": false,
  "env_version": "release"
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| scene | 小程序码携带参数，使用短字段 `v` |
| page | 小程序落地页；当前后端配置为 `pages/user-sub/verifier-bind/index`，该页面需要解析 `scene` 并进入核销员绑定流程 |
| check_path | 设为 `false`，避免未发布页面在生成时被微信拦截 |
| env_version | 默认生成正式版小程序码 |

### 7.3 图片存储

微信接口返回的是图片二进制内容，不是 URL。

后端拿到图片后上传 OSS：

```text
verifier/qrcode/YYYY/MM/DD/<verifier_id>.png
```

最终返回给后台前端：

```json
{
  "wechat_mini_program_code_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
  "wechat_qr_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png"
}
```

后台页面直接展示 `wechat_mini_program_code_url` 或 `wechat_qr_url` 即可。`wechat_qr`、`wechat_qr_url`、`wechat_mini_program_code_url` 三个字段当前都返回同一张微信小程序码图片 URL。

### 7.4 依赖配置

生成微信小程序码依赖：

| 配置 | 说明 |
| --- | --- |
| app.app_id | 微信小程序 AppID |
| app.app_secret | 微信小程序 AppSecret |
| oss | OSS 上传配置 |

如果微信配置错误，后端会返回微信错误信息，例如 `微信 access_token 获取失败` 或 `微信小程序码生成失败`。

## 8. 前端联调注意点

- 后台添加核销员时只提交 `name` 和 `phone`。
- 后台页面优先展示 `wechat_mini_program_code_url` 或 `wechat_qr_url`。
- 小程序激活页读取微信 `scene`，从 `v` 映射出 `verifier_id`，`channel` 固定为 `wechat`。
- - 小程序必须先确保用户已登录/注册，再调用激活接口。
- 激活接口请求体里的 `phone` 应使用当前登录/注册手机号。
- 如果用户扫码后手机号不一致，应提示用户使用后台登记的手机号登录。

## 9. 扫码空白页排查

如果微信可以扫出小程序，但进入后页面是空白，通常不是后端二维码图片坏了，而是小程序端落地页没有准备好。

需要确认：

1. 小程序项目里存在页面 `pages/user-sub/verifier-bind/index`。
2. `app.json` 已注册该页面。
3. 当前扫码打开的版本包含这个页面。后端默认生成 `env_version = release`，所以正式码会打开已发布版本。
4. 页面 `onLoad` 已解析微信传入的 `scene`。

小程序页面最小示例：

```js
Page({
  onLoad(options) {
    const scene = decodeURIComponent(options.scene || "")
    const params = {}
    scene.split("&").forEach((item) => {
      const [key, value] = item.split("=")
      if (key) params[key] = value
    })

    const verifierId = params.v
    const channel = "wechat"

    if (!verifierId) {
      wx.showToast({ title: "核销员码无效", icon: "none" })
      return
    }

    // 这里进入登录/注册流程；登录成功后调用：
    // POST /api/v1/verifier/activate
    // { verifier_id: Number(verifierId), phone, channel }
  }
})
```

如果小程序页面还没发布，可以临时把后端生成参数里的 `env_version` 改为 `trial` 或 `develop`，但正式上线应使用 `release`。

## 10. 当前仍保留的临时点

核销订单相关接口当前仍使用：

```http
X-Verifier-Id: <verifier_id>
```

例如：

```http
POST /api/v1/verifier/confirm
GET /api/v1/verifier/verified-list
```

后续可以继续升级为“通过登录用户 `user_id` 查询已绑定的核销员身份”，从而去掉 `X-Verifier-Id`。
