# 2026-08-12 产品反馈 11 项前端对接总说明

本文对应今日确认的 11 个产品问题，供小程序、商家端和管理端前端统一对接。除特别说明外，基础前缀均为：

```text
https://www.hypercn.cn/api/v1
```

需要登录的接口统一携带：

```http
Authorization: Bearer <access_token>
```

## 一览

| # | 功能点 | 后端状态 | 前端动作 |
|---:|---|---|---|
| 1 | 入驻分类、场地发布流程 | 已支持 | 入驻选择类型；场地发布移除全部票务 UI |
| 2 | 同一商家多账号身份展示 | 已支持 | 商家页统一使用 `organizer` 身份 |
| 3 | 个人主页 Logo 居中 | 纯前端 | 调整头部布局 |
| 4 | 客服消息与管理端客服工作台 | 已支持 | 小程序发起咨询，管理端接入工作台 |
| 5 | 他人主页关注/粉丝误跳自己 | 纯前端 | 禁用或隐藏入口 |
| 6 | 我的赞、我的收藏 | 已支持 | 个人主页新增两个 Tab |
| 7 | 已入驻商家发布内容/公开主页 | 已支持 | 商家端用统一组织范围查询；C 端接商家主页 |
| 8 | 发布活动/场地选择优惠标签 | 已支持 | 获取标签并提交 `tag_ids` |
| 9 | 首页搜索出现下架/过期内容 | 已支持 | 统一走新搜索/地图接口并按 ID 契约跳转 |
| 10 | 购票页封面放大与活动详情 | 已支持字段 | 前端展示海报、长图、详情正文 |
| 11 | 底部四 Tab 白边闪烁 | 纯前端 | 修正页面背景与切换动画 |

---

## 1. 入驻选择类型，场地发布不得出现票务

### 1.1 入驻申请

```http
POST /organizer/apply
```

```json
{
  "name": "Hyper",
  "type": "venue",
  "logo": "https://cdn.hypercn.cn/organizer/logo.png",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区"
}
```

| `type` | 页面初始业务类型 |
|---|---|
| `venue` | 场地 |
| `party` | 派对 |

该字段用于入驻申请与审核分类，**不限制同一商家之后只能发布场地或派对之一**。

入驻申请页只提交文档允许的资料，不提交票种、票价、库存、售卖时间、实名配置等票务字段。

### 1.2 发布场地

```http
POST /activity/create
```

```json
{
  "activity_id": 0,
  "step": 1,
  "type": "venue",
  "name": "Hyper Space",
  "description": "场地介绍",
  "business_hours": "每天 19:30-次日 02:30",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区",
  "address": "天府三街",
  "latitude": 30.657,
  "longitude": 104.066,
  "tag_ids": [101, 102]
}
```

前端必须遵守：

- `type=venue`：不展示票种、库存、价格、售卖起止时间、实名、支付或购票相关步骤。
- 场地只展示和提交 `business_hours`，例如“每天 19:30-次日 02:30”。
- `type=party`：保留活动起止时间、票种和售卖配置。
- 场地不能调用 `POST /activity/:id/ticket-specs`，也不能在 `/activity/create` 传非空 `ticket_specs`。后端会拒绝并返回“场地不支持票券配置”。
- 场地的底层兼容有效期不对 C 端展示，不要把 `start_time`、`end_time` 当营业时间。

---

## 2. 同一商家多账号统一显示商家身份

### 本人信息

```http
GET /user/info
```

商家所有者或已启用的商家子账号会收到：

```json
{
  "organizer": {
    "id": 7,
    "type": "merchant",
    "name": "Hyper",
    "logo": "https://cdn.hypercn.cn/organizer/logo.png"
  }
}
```

前端约定：

- 个人资料页继续使用 `user.nickname`、`user.avatar_url`。
- 商家后台头部、商家发布内容、商家卡片、商家主页使用 `organizer.name`、`organizer.logo`。
- 不要用创建活动的个人昵称和头像覆盖商家名和 Logo。

已启用的 `organizer_staff` 子账号调用以下接口，会自动归属同一组织：

```http
GET /organizer/info
GET /activity/my-list
POST /activity/create
GET /organizer/orders
```

### 商家发布内容列表

```http
GET /activity/my-list?page=1&size=20&status=3
```

- `status=3`：已发布/已上线内容。
- 草稿、待审核、驳回应按状态单独查询和展示，不要在“已发布”页混入。
- 子账号和商家创建者都应看到同一组织的内容。

---

## 3. 个人主页顶部 Logo 居中

此项没有后端改动。

- 商家 Logo 容器采用真正的居中布局，不能被左侧返回按钮或右侧操作按钮挤偏。
- 推荐头部使用三列固定布局：左操作区、绝对居中的品牌区、右操作区。
- 不要用文字宽度或左右元素占位差异模拟居中。

---

## 4. 客服消息：小程序咨询 + 管理端客服工作台

### 4.1 小程序端获取客服身份

```http
GET /user/customer-service
```

当前部署配置的客服用户为 `users.id=77`，但**小程序不得写死 `77`**，必须先请求本接口。

```json
{
  "code": 200,
  "data": {
    "user_id": 77,
    "nickname": "Hyper 客服",
    "avatar_url": "https://cdn.hypercn.cn/service.png",
    "signature": "在线客服"
  }
}
```

接口返回 `404` 时，展示“客服暂不可用”。

### 4.2 小程序端发送咨询

```http
POST /message/send
```

```json
{
  "target_id": "77",
  "session_type": 1,
  "msg_type": 1,
  "content": "你好，我需要帮助"
}
```

消息历史仍使用既有 IM：

```http
GET /message/list?peer_id=77&session_type=1&cursor=0
POST /session/clear-unread
```

### 4.3 管理端 PC 客服工作台

菜单建议：`消息管理 > 客服工作台`。管理员登录后台后使用以下接口，权限码为 `admin.customer_service`。

```http
GET  /admin/customer-service/sessions?page=1&pageSize=20&keyword=
GET  /admin/customer-service/sessions/:user_id/messages?cursor=0&pageSize=20
POST /admin/customer-service/sessions/:user_id/messages
POST /admin/customer-service/sessions/:user_id/read
```

回复示例：

```json
{
  "msg_type": 1,
  "content": "您好，已为您处理。",
  "parent_msg_id": "0",
  "ext": {}
}
```

PC 页面约定：

- 会话列表使用 `GET /admin/customer-service/sessions`；可按客户昵称、手机号、数值用户 ID 搜索。
- 点击会话后请求历史，并调用已读接口。
- 聊天气泡按历史消息 `is_self` 布局：`true` 是平台客服发出的消息，`false` 是客户。
- 管理员回复后，客户端用户只看到“Hyper 客服”，不会看到操作管理员名称。
- 客服回复和标记已读会记录管理员操作日志。
- 管理员只能回复用户已经发起的咨询，不能借此接口主动私信任意用户。

完整字段、错误码和上线配置见：[admin_customer_service_workbench_api_20260812.md](/Users/luosmallrui/Hyper/docs/admin_customer_service_workbench_api_20260812.md)。

---

## 5. 查看他人主页时禁用关注/粉丝列表入口

此项没有后端改动。

当前关注、粉丝列表接口的查询主体是当前登录用户，并不是资料页目标用户。为避免打开别人主页却跳到自己列表：

- 当 `profile.user.id !== currentUser.id` 时，隐藏或禁用“关注”和“粉丝”可点击入口。
- 保留数字展示即可，暂不做跳转。
- 不要把当前用户列表接口伪装成他人社交关系列表。

---

## 6. 个人主页新增“赞过”“收藏” Tab

```http
GET /note/my/likes?page=1&pageSize=20
GET /note/my/collects?page=1&pageSize=20
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "notes": [],
    "total": 0
  }
}
```

前端约定：

- 在“我的动态”右侧新增“赞过”“收藏”两个 Tab，仅在本人主页展示。
- `notes` 沿用现有帖子卡片渲染结构。
- 分页参数使用 `page`、`pageSize`。
- 后端允许收藏自己的公开帖子，前端不需要禁止。

---

## 7. 已入驻商家的公开主页与内容展示

### C 端商家主页

```http
GET /organizers/:id?activity_page=1&activity_size=10&venue_page=1&venue_size=10
```

`id` 是活动详情中的 `organizer.id`，不是创建者 `user_id`。

响应同时提供：

- `name`、`logo`、`cover_image`、`gallery`、`description`、地址、客服电话、人均消费；
- `activities.list`：该商家已上架、未结束的派对；
- `venues.list`：该商家已上架场地；
- `business_hours`：场地每日营业时间；
- `follow_count`、`is_follow`：商家主页独立关注状态。

点击规则：

- 派对卡片：`GET /activity/:id`。
- 场地卡片：`GET /venues/:id`，这里 `id=organizer_id`。
- 公开主页不展示审核资料、银行卡或后台入口。

---

## 8. 发布活动/场地时选择优惠标签

### 获取标签

```http
GET /content-tags
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {"id": 101, "name": "积分立减", "image": "", "value": "", "sort": 1},
      {"id": 102, "name": "买单立减", "image": "", "value": "", "sort": 2},
      {"id": 103, "name": "新人优惠", "image": "", "value": "", "sort": 3}
    ]
  }
}
```

前端必须动态读取标签，不能写死标签 ID。标签由管理端配置，主办方只可多选。

### 保存标签

在原 `POST /activity/create` 请求中携带：

```json
{
  "activity_id": 10,
  "step": 1,
  "type": "party",
  "tag_ids": [101, 103]
}
```

- 字段不传：不修改已绑定标签。
- `tag_ids: []`：清空标签。
- 派对标签绑定活动；场地标签绑定场地主办方，场地详情、地图及筛选会使用同一组标签。
- 无效、删除或停用的标签 ID 会被后端拒绝。

---

## 9. 首页搜索和地图不展示下架、未审核、过期内容

### 搜索

```http
GET /search/?keyword=hyper&type=0&tag_ids=101&lat=30.657&lng=104.066
```

`type`：

| 值 | 内容 |
|---:|---|
| `0` | 综合 |
| `1` | 用户 |
| `2` | 帖子 |
| `3` | 新场地 |
| `4` | 未结束的新派对/票务活动 |

后端会过滤：

- 主办方未审核或停用；
- `activities.status != 3`；
- `is_hidden=1`；
- 已结束的派对。

### 地图

```http
GET /map/markers?source=all&keyword=hyper&tag_ids=101
```

前端跳转必须使用 `detail_url`，或使用 `source + source_id`，不要自行从字符串 `id` 截取数字。

| 内容 | `source` | `source_id` | 详情 |
|---|---|---:|---|
| 场地 | `venue` | `organizer_id` | `/venues/:source_id` |
| 派对 | `activity` | `activity_id` | `/activity/:source_id` |

场地每日营业结束也仍是场地，不应因为当天结束营业而从地图消失。

---

## 10. 购票页封面预览与活动完整详情

继续调用：

```http
GET /activity/:id
```

前端展示字段：

| 字段 | 用途 |
|---|---|
| `poster_list` | 购票页封面，点击后用图片预览组件放大 |
| `poster_detail` | 活动详情海报 |
| `poster_long` | 长图详情海报，按原比例完整展示 |
| `poster_wechat` | 微信分享海报 |
| `description` | 活动正文/富文本详情 |
| `organizer` | 主办方名称、Logo、主页跳转 ID |
| `tags` | 优惠标签 |
| `ticket_specs` | 派对票种；场地固定为空数组 |

场地详情不得显示购买票券区域；票务活动才展示购票和票种。

---

## 11. 底部四 Tab 切换白边闪烁

此项没有后端改动。

前端排查与修复要求：

1. 四个根页面、页面容器和自定义 TabBar 使用同一背景色。
2. 避免页面卸载瞬间暴露默认白色 `body` / 根节点背景。
3. 切换时不要插入白色 loading 占位层；骨架屏或 loading 容器继承当前页面背景。
4. 清理导致 1px 白边的 `border`、`safe-area`、透明过渡层和缩放动画。
5. 在真机连续快速切换四个 Tab 验收，不只在开发者工具验证。

---

## 统一验收清单

1. 场地发布全流程无票务字段或票务页面；派对发布保留票务流程。
2. 两个商家子账号登录后均显示相同商家 Logo、名称和发布内容。
3. 本人主页可切换动态、赞过、收藏；他人主页关注/粉丝入口不可跳转。
4. 小程序用户可向 `/user/customer-service` 返回的账号咨询；管理端客服工作台能看到、已读并回复。
5. 发布页标签来源于 `/content-tags`；标签可保存、详情可展示、地图和搜索可筛选。
6. 下架或结束派对不再出现在搜索和地图结果；场地仍可展示。
7. 活动购票页可放大封面并完整展示详情海报、长图和正文。
8. 真机切换底部 Tab 无白边闪烁。

## 关联文档

- [客服工作台完整接口文档](admin_customer_service_workbench_api_20260812.md)
- [活动、场地与派对标签接口](content_tag_management_api_20260807.md)
- [C 端商家主页接口](public_organizer_home_api_20260811.md)
- [本轮产品反馈后端改动](client_product_feedback_api_20260812.md)
