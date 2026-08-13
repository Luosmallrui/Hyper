# 后端待修复事项处理结果与前端验收说明

更新时间：2026-08-13

本文回应前端提出的 `backend_fix_requests_20260810(1).md`。除明确标记为“待部署/待配置”的内容外，接口代码均已完成；接口 Base URL 为 `/api/v1`。

## 1. 处理结论

| 项目 | 状态 | 前端动作 |
|---|---|---|
| 商家活动列表状态语义 | 已支持 | 使用 `status=0~4` 筛选；不传 `status` 时展示全部状态。 |
| 商家子账号查看同一组织数据 | 已支持 | 子账号完成 `organizer_staff` 绑定且状态启用后，继续使用原接口。 |
| 客服聊天账号 | 待生产配置 | 客户端继续调用 `GET /user/customer-service`；后台将客服账号配置为用户 ID `77` 后生效。 |
| 活动/场地/派对对象关注 | 待建表并部署 | 前端按对象关注文档传 `target_type`、`target_id`。 |
| `GET /organizer/info` 参数报错 | 源码已修复，待部署验证 | 部署后以主办方和子账号 Token 各测一次。 |
| 订单关联活动被删除 | 已修复兼容逻辑，待部署 | 依据 `activity.is_hidden` 显示下架态，不要把空海报视为接口失败。 |
| 密码登录 Token 续期字段 | 已支持 | 按 `access_expire` 和 `refresh_expire` 管理 Token。 |
| 活动详情 `user_id` | 已支持 | 直接读取 `data.user_id`，不要自行推导。 |
| 活动 PV/UV、转化率 | 已支持，待建表并部署 | 游客详情请求带稳定的 `X-Visitor-Id`。 |
| 销售渠道、提现状态筛选 | 已支持 | 商家订单列表传对应查询参数。 |

## 2. 历史订单活动已物理删除

已确认数据库 `activities` 表中不存在以下活动 ID：

```text
8
9
10
13
```

而已有订单仍引用这些 ID，例如订单 `T202606140031319bb9eec9` 关联活动 `8`，多笔订单关联活动 `9`、`10`、`13`。

因此无法从现有订单表反推出原活动名称、时间和海报。若需要恢复真实原始内容，只能从数据库备份恢复相应 `activities` 行及必要的 `ticket_specs` 数据。

### 2.1 历史订单接口兼容结果

部署本次后端后，以下接口不会再因为活动不存在返回 `500`：

```http
GET /api/v1/order/list?page=1&size=50
GET /api/v1/order/:order_no
GET /api/v1/activity/:id
```

对于关联已删除活动的订单，订单列表和订单详情使用既有字段返回下架兜底：

```json
{
  "activity": {
    "id": 8,
    "name": "活动已下架",
    "poster_list": "",
    "is_hidden": true,
    "hidden_reason": "活动已下架"
  }
}
```

约定：

- `is_hidden=true` 是唯一需要依赖的下架判断字段。
- `poster_list` 为空时展示客户端默认封面，不请求或猜测历史图片地址。
- 下架活动的开始、结束时间可能为零值；前端不应格式化为 `0001-...`，建议展示 `-`。
- 已购票用户打开 `GET /activity/:id` 时可获得只读下架兜底详情；未购票用户和游客仍返回 `404`，防止已删除内容重新公开。
- 下架详情不提供票种、不允许订阅、关注或下单。

### 2.2 后续删除活动的规则

接口路径和调用方式不变：

```http
DELETE /api/v1/activity/:id
```

后端规则调整为：

1. 活动没有任何订单时，仍物理删除。
2. 活动已有任意订单时，不再物理删除，而是更新为：

```json
{
  "is_hidden": true,
  "hidden_reason": "主办方已下架活动"
}
```

3. 下架活动不会出现在公开地图、搜索、推荐和购票入口；历史订单、退款、核销、对账仍可保留关联信息。

前端无需改调用接口。删除成功后，将商家端列表内该活动标记为“已下架”即可。

## 3. 商家活动列表与子账号

### 3.1 活动状态

```http
GET /api/v1/activity/my-list?page=1&size=20&status=3
Authorization: Bearer <organizer_or_staff_access_token>
```

| `status` | 含义 |
|---:|---|
| `0` | 草稿 |
| `1` | 待提交/待审核 |
| `2` | 审核中 |
| `3` | 已上架 |
| `4` | 已驳回 |

- 不传 `status`：返回当前商家全部活动，不是只返回已上架活动。
- `keyword`、`published_from`、`published_to`、`activity_from`、`activity_to` 筛选仍可组合使用。

### 3.2 子账号

已启用的 `organizer_staff.user_id` 会被解析为对应的 `organizer_id`，因此子账号可读取同一商家的活动、订单、资料和统计数据。若子账号仍返回空列表，请后端运维确认：

```sql
SELECT id, organizer_id, user_id, role_id, name, phone, status
FROM organizer_staff
WHERE user_id = <子账号用户ID>;
```

要求 `status = 1`，且对应 `organizers.status = 2`、`organizers.enabled = 1`。

## 4. 客服账号配置

客户端接口保持不变：

```http
GET /api/v1/user/customer-service
Authorization: Bearer <access_token>
```

生产库应配置已确认的客服用户 `users.id = 77`：

```sql
INSERT INTO platform_settings (setting_key, setting_value, remark)
VALUES ('customer_service_user_id', '77', '客服聊天用户 ID')
ON DUPLICATE KEY UPDATE
  setting_value = VALUES(setting_value),
  remark = VALUES(remark);
```

配置成功后，客户端使用响应的客服用户 ID 创建或打开既有单聊会话。客服在管理端工作台处理会话的接口见 [admin_customer_service_workbench_api_20260812.md](admin_customer_service_workbench_api_20260812.md)。

## 5. 对象关注

活动、场地、旧派对的关注与“关注用户”已分离。前端在原接口请求体中增加对象参数：

```http
POST /api/v1/follow/follow
POST /api/v1/follow/unfollow
```

```json
{
  "target_type": "activity",
  "target_id": 15
}
```

列表和详情响应使用：

```json
{
  "is_follow": true,
  "follow_count": 28,
  "follow_target_type": "activity",
  "follow_target_id": 15
}
```

前端必须直接使用服务端返回的 `follow_target_type` 和 `follow_target_id`，不要用 `user_id`、`source_id` 推断关注对象。完整约定见 [content_follow_api_20260810.md](content_follow_api_20260810.md)。

生产库部署前必须创建 `content_follows` 表，见该文档第 5 节 SQL。

## 6. 登录与活动详情

### 6.1 密码登录

```http
POST /api/v1/auth/login-password
```

成功响应已包含：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expire": 1780000000,
  "refresh_expire": 1780600000
}
```

客户端在 access token 即将过期或收到鉴权失败后，调用：

```http
POST /api/v1/auth/refresh
```

### 6.2 活动详情

```http
GET /api/v1/activity/:id
```

正常活动详情已返回：

```json
{
  "user_id": 35,
  "is_follow": false,
  "is_subscribe": false
}
```

`user_id` 是活动主办方用户 ID。

## 7. 商家主页身份字段

商家主页接口：

```http
GET /api/v1/organizers/:id
```

页面头部必须展示商家实体信息：

```json
{
  "id": 7,
  "name": "Hyper",
  "logo": "https://cdn.example.com/organizer-logo.png"
}
```

字段约定：

| 页面位置 | 必须使用 | 禁止使用 |
|---|---|---|
| 主页头像/Logo | `data.logo` | `data.owner_avatar` |
| 主页名称 | `data.name` | `data.owner_nickname` |
| 个人资料、动态作者等个人身份位置 | `owner_nickname`、`owner_avatar` 可按页面需要展示 | 不得覆盖商家头部 |

`owner_nickname`、`owner_avatar` 是入驻申请人/主办方账号的个人资料，仅供明确标注“主理人”或动态作者时使用，不是商家品牌资料。商家名称和 Logo 可由主办方在 `PUT /api/v1/organizer/profile` 或 `PUT /api/v1/organizer/basic` 维护。

## 8. 浏览统计、渠道与提现状态

### 7.1 PV/UV 与转化率

公开活动详情成功读取后，后端自动计 PV。游客端应为同一设备长期保存一个 UUID，并在每次详情请求携带：

```http
X-Visitor-Id: <stable-uuid>
```

商家统计接口：

```http
GET /api/v1/activity/:id/statistics
GET /api/v1/activity/:id/statistics/daily?days=7
```

生产库须执行 `activity_daily_stats`、`activity_daily_visitors` 和 `organizer_withdraw_allocations` 建表 SQL，详见 [activity_traffic_withdraw_status_api_20260810.md](activity_traffic_withdraw_status_api_20260810.md)。

### 7.2 商家订单筛选

```http
GET /api/v1/organizer/orders?page=1&size=20&sales_channel=wechat&withdraw_status=available
```

支持：

| 参数 | 可选值 |
|---|---|
| `sales_channel` | `wechat`、`douyin`、`web`、`other` |
| `withdraw_status` | `available`、`pending_withdraw`、`withdrawn` |
| `start_date` | `YYYY-MM-DD` |
| `end_date` | `YYYY-MM-DD`，包含当天 |

完整字段说明见 [sales_channel_order_filter_api_20260810.md](sales_channel_order_filter_api_20260810.md)。

## 9. 上线顺序与验收

1. 生产库执行本轮 `content_follows`、浏览统计和提现分配表的建表 SQL，并配置 `customer_service_user_id=77`。
2. 部署后端版本。
3. 使用主办方主账号和子账号分别验证 `GET /organizer/info`、`GET /activity/my-list`。
4. 验证历史订单 `T202606140031319bb9eec9`：订单列表和详情均应返回 `activity.is_hidden=true`，不得出现 `500` 或“活动不存在”。
5. 创建一条测试活动并下单，再调用 `DELETE /activity/:id`：活动应在公开端消失，历史订单仍能查看且标记下架。
6. 使用两个不同用户验证对象关注后，重新请求地图或详情，确认 `is_follow` 和 `follow_count` 变化。
7. 请求 `GET /organizers/:id`，确认页面头部使用 `name`、`logo`，而非 `owner_nickname`、`owner_avatar`。
8. 游客连续打开同一活动详情，确认请求携带同一个 `X-Visitor-Id`；商家侧验证 PV、UV 与转化率。
