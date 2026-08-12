# 产品反馈后端接口更新

更新时间：2026-08-12

本文说明本轮产品反馈中已由后端处理的接口契约，以及需要小程序/网页前端直接调整的 UI 行为。

## 1. 入驻类型与发布流程

### 1.1 入驻时选择类型

```http
POST /api/v1/organizer/apply
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求新增/启用 `type`：

```json
{
  "name": "Hyper",
  "type": "venue",
  "logo": "https://cdn.example.com/hyper.png",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区"
}
```

| 值 | 含义 |
|---|---|
| `venue` | 场地 |
| `party` | 派对。后端兼容保存为历史值 `merchant` |

`type` 用于申请分类和后台审核展示，**不锁定主办方后续只能发布一种内容**。同一商家仍可以发布场地和派对。

入驻申请不提交票种、库存、售价、售卖时间等票务字段。

### 1.2 创建场地时不得显示或提交票务步骤

```http
POST /api/v1/activity/create
Authorization: Bearer <access_token>
```

场地示例：

```json
{
  "activity_id": 0,
  "step": 1,
  "type": "venue",
  "name": "Hyper Space",
  "description": "场地介绍",
  "business_hours": "19:30-02:30",
  "address": "天府三街",
  "latitude": 30.657,
  "longitude": 104.066,
  "tag_ids": [101, 102]
}
```

前端规则：

- `type=venue`：不展示票种、库存、实名、售卖起止时间、支付等任何票务 UI，不调用票种保存接口。
- `type=party`：保留既有活动时间和票务配置步骤。
- 场地可填写 `business_hours`，派对不可填写该字段。
- 后端会拒绝场地的 `step=4` 或任何非空 `ticket_specs`，错误信息为“场地不支持票券配置”。
- 已有票种的派对不能转换成场地，需新建场地，避免破坏既有订单。

`GET /api/v1/activity/:id/ticket-specs` 对场地返回空数组；`POST /api/v1/activity/:id/ticket-specs` 对场地返回业务错误。

## 2. 商家身份与多账号

### 2.1 本人资料中的商家身份

```http
GET /api/v1/user/info
Authorization: Bearer <access_token>
```

当登录用户是审核通过且启用的商家所有者，或该商家的启用子账号时，响应新增：

```json
{
  "organizer": {
    "id": 7,
    "type": "merchant",
    "name": "Hyper",
    "logo": "https://cdn.example.com/hyper-logo.png"
  }
}
```

`user` 仍是个人资料；进入商家后台、商家卡片或商家主页时必须使用 `organizer.name`、`organizer.logo`，不要用个人 `user.nickname`、`user.avatar_url` 替代。

### 2.2 商家端接口支持子账号归属

已启用的 `organizer_staff` 子账号调用以下商家接口时，都会解析到同一个组织：

```http
GET /api/v1/organizer/info
GET /api/v1/activity/my-list
POST /api/v1/activity/create
GET /api/v1/organizer/orders
```

因此两个账号被加入同一商家后，读取到的是同一个商家名称、头像、活动列表和订单范围。

已有活动列表继续调用：

```http
GET /api/v1/activity/my-list?page=1&size=20
Authorization: Bearer <access_token>
```

已发布内容筛选：`status=3`。草稿/待审核也应按状态独立展示，不能把“发布列表”误做成仅查询个人用户 ID。

### 2.3 地图与搜索中的商家展示

新活动/场地地图标记已使用商家名和商家 Logo：

```json
{
  "user": "Hyper",
  "username": "Hyper",
  "user_avatar": "https://cdn.example.com/hyper-logo.png",
  "userAvatar": "https://cdn.example.com/hyper-logo.png"
}
```

前端不要再以创建者用户资料覆盖这些字段。

## 3. 发布优惠标签

### 3.1 获取可选标签

```http
GET /api/v1/content-tags
```

匿名和登录状态均可调用。只返回管理端启用的 `coupon_tag`：

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

标签由管理端配置，主办方只能多选现有标签，不能自由录入。

### 3.2 保存标签

创建或编辑时，将标签 ID 放在现有 `/activity/create` 请求中：

```json
{
  "activity_id": 10,
  "step": 1,
  "type": "party",
  "tag_ids": [101, 103]
}
```

- 字段未传：不改变已有标签。
- `tag_ids: []`：清空全部标签。
- 派对标签绑定到活动；场地标签绑定到场地主办方，场地详情、地图和场地筛选会共享同一组标签。
- 无效、已删除或已停用的 ID 会返回错误。

详情与列表继续读取 `tag_ids`、`tags`，地图还兼容 `discount_tags`。

## 4. 首页搜索、地图与下架过滤

### 4.1 搜索

```http
GET /api/v1/search/?keyword=hyper&type=0&tag_ids=101&lat=30.657&lng=104.066
```

`type` 约定：

| type | 数据 |
|---|---|
| `0` | 综合：用户、帖子、新场地、未结束的新派对 |
| `1` | 用户 |
| `2` | 帖子 |
| `3` | 新场地 |
| `4` | 未结束的新派对/票务活动 |

搜索已不再读取旧 `merchants` 表。新内容必须同时满足：

- 主办方审核通过且已启用；
- `activities.status=3`；
- `is_hidden=0`；
- 派对还必须满足 `end_time >= 当前时间`。

场地结果位于 `data.parties` 以保持旧字段兼容，但其 ID 语义为：

```json
{
  "id": 7,
  "activity_id": 31,
  "type": "venue"
}
```

- `id`：场地 `organizer_id`，进入场地详情时调用 `GET /api/v1/venues/7`。
- `activity_id`：承载该场地展示信息的活动记录，仅用于内部追踪，不要用它调用场地详情。

### 4.2 地图

```http
GET /api/v1/map/markers?source=all&keyword=hyper&tag_ids=101
```

新场地标记：

```json
{
  "id": "venue-7",
  "source": "venue",
  "source_id": 7,
  "detail_type": "venue",
  "detail_url": "/api/v1/venues/7"
}
```

派对标记仍为 `source=activity`、`source_id=activity_id`、详情 `/api/v1/activity/:id`。前端应优先使用 `detail_url` 或 `source + source_id` 跳转，不要用字符串 `id` 自行截取。

地图同样会过滤下架、未审核、已禁用商家和已结束派对；场地不因每日营业结束而从地图消失。

## 5. 个人主页“我的赞”和“我的收藏”

```http
GET /api/v1/note/my/likes?page=1&pagesize=20
GET /api/v1/note/my/collects?page=1&pagesize=20
Authorization: Bearer <access_token>
```

统一响应：

```json
{
  "code": 200,
  "data": {
    "notes": [],
    "total": 0
  }
}
```

- `my/likes`：当前用户点过赞的公开帖子，按最近点赞时间倒序。
- `my/collects`：当前用户收藏的公开帖子。
- 建议个人主页在“我的动态”右侧新增“赞过”和“收藏”两个 tab，登录本人时才展示。
- 后端允许收藏自己的公开帖子，客户端无需限制。

## 6. 客服私聊

### 6.1 先获取客服账号

```http
GET /api/v1/user/customer-service
Authorization: Bearer <access_token>
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "user_id": 52,
    "nickname": "Hyper 客服",
    "avatar_url": "https://cdn.example.com/customer-service.png",
    "signature": "工作日 10:00-22:00"
  }
}
```

未配置、账号不存在或已停用时返回 `404`，客户端展示“客服暂不可用”，不要写死客服用户 ID。

### 6.2 复用现有 IM 单聊

```http
POST /api/v1/message/send
```

```json
{
  "target_id": "52",
  "session_type": 1,
  "msg_type": 1,
  "content": "你好，我需要帮助"
}
```

客服端使用相同 IM 会话列表和消息列表接口接收。先从会话列表取得咨询用户 ID，再以该用户为 `peer_id` 拉取历史：

```http
GET /api/v1/session/list
GET /api/v1/message/list?peer_id=<咨询用户ID>&session_type=1
```

部署后需由管理员在 `PUT /api/v1/admin/system-config` 配置 `customer_service_user_id`，其值必须是一个正常状态的用户 ID。客服账号不在线时，消息仍会入库并出现在该账号后续登录的会话列表中。

## 7. 购票详情已有字段

购票页无需新增后端接口，活动详情已经提供：

```http
GET /api/v1/activity/:id
```

前端可使用：

- `poster_list`：封面图，点击用图片预览组件放大；
- `poster_detail`：详情海报；
- `poster_long`：长图海报；
- `description`：活动正文；
- `organizer`：主办方信息；
- `ticket_specs`：仅派对票务活动返回，场地返回空数组；
- `tags`：优惠标签。

## 8. 仅前端处理项

以下问题没有后端 API 缺口，应由前端直接修复：

1. 个人主页顶部 Hyper Logo 的水平居中。
2. 查看他人个人主页时，关注/粉丝入口直接禁用或隐藏；现有列表接口是“当前登录用户”的列表，不能拿它展示别人。
3. 购票页封面点击预览、详情海报/长图展示、文案布局。
4. 四个底部 Tab 切换时的白边闪烁：检查页面背景色、根容器背景和转场/销毁时机，保持与导航背景一致。

## 9. 数据库与部署

本轮业务接口不需要新建表或新增列。客服账号使用既有 `platform_settings` 表保存，配置接口第一次提交正数 `customer_service_user_id` 时会自动写入。

部署后建议按以下顺序验收：

1. 管理端配置客服用户 ID。
2. 用两个已绑定同一 `organizer_staff` 的账号分别调用 `/organizer/info`、`/activity/my-list`。
3. 创建一个 `venue`，确认票券步骤不出现且保存票种被拒绝。
4. 给场地与派对各选择标签，验证 `/map/markers`、`/search/` 的 `tag_ids` 筛选。
5. 下架活动、结束派对后，验证首页地图和搜索均不再返回。
