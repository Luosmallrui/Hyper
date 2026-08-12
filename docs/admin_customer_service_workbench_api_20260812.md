# 管理端站内客服工作台接口

更新时间：2026-08-12

## 目标与边界

本接口为管理端 PC 提供站内 IM 客服工作台。它复用已有单聊消息、会话未读、RocketMQ 消息投递和 Socket 实时推送链路，不创建第二套聊天数据。

- 小程序用户统一与平台客服账号聊天。
- 管理员从 PC 工作台查看和回复客服会话。
- 用户看到的回复发送者始终是平台客服账号，不暴露具体管理员个人身份。
- 管理员实际回复、标记已读操作会写入管理员操作日志。
- 工作台只读取“平台客服账号”和用户之间的单聊；管理员不能通过本接口查看任意用户私信，也不能主动给从未咨询过的用户发送客服消息。
- 本期不接入微信原生客服或微信订阅消息；后续接入时，仍以站内 IM 为聊天记录主数据。

基础前缀：`/api/v1/admin`

认证：

```http
Authorization: Bearer <admin_access_token>
```

RBAC 权限：`admin.customer_service`

## 上线配置

先准备一个正常状态的 `users` 账号，作为对小程序用户展示的“Hyper 客服”账号。然后由超级管理员在系统配置中设置其用户 ID：

```http
PUT /api/v1/admin/system-config
```

```json
{
  "system_name": "Hyper",
  "icp_record_no": "蜀ICP备2026000362号",
  "customer_service_phone": "",
  "customer_service_wechat": "",
  "customer_service_email": "",
  "customer_service_hours": "",
  "customer_service_user_id": 52,
  "withdraw_arrival_cycle": "T+1 到 T+3 个工作日"
}
```

`customer_service_user_id` 必须是正常状态用户。未配置、值为 `0`、账号不存在或账号被停用时，客服工作台接口返回：

```json
{
  "code": 409,
  "msg": "客服工作台未配置客服账号"
}
```

已有数据库需执行：

```sql
INSERT IGNORE INTO platform_settings (setting_key, setting_value, remark)
VALUES ('customer_service_user_id', '0', '客服聊天用户 ID');
```

## 1. 客服会话列表

```http
GET /api/v1/admin/customer-service/sessions?page=1&pageSize=20&keyword=13800000000
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---:|---|
| `page` | 否 | 页码，默认 `1` |
| `pageSize` | 否 | 每页数量，默认 `20`，最大 `100` |
| `keyword` | 否 | 客户昵称、手机号或数值用户 ID |

成功响应：

```json
{
  "code": 200,
  "data": {
    "service_user_id": 52,
    "list": [
      {
        "user_id": 35,
        "nickname": "农子健",
        "avatar": "https://cdn.hypercn.cn/avatars/35.png",
        "mobile": "13800000000",
        "user_status": 1,
        "last_msg_id": "2060000000000000000",
        "last_msg_type": 1,
        "last_msg": "退款什么时候到账？",
        "last_msg_time": 1786500000000,
        "unread": 1
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

列表按置顶、最后消息时间倒序返回。`unread` 是平台客服账号未读数，不是客户未读数。

## 2. 获取客服聊天记录

```http
GET /api/v1/admin/customer-service/sessions/35/messages?cursor=0&pageSize=20
```

路径参数：

| 参数 | 说明 |
|---|---|
| `user_id` | 客户数值用户 ID，即会话列表的 `user_id` |

查询参数：

| 参数 | 必填 | 说明 |
|---|---:|---|
| `cursor` | 否 | 上翻历史消息的毫秒时间戳；首次传 `0` 或不传 |
| `since` | 否 | 拉取此时间点之后的新消息，毫秒时间戳 |
| `pageSize` | 否 | 条数，默认 `20`，最大 `100` |

成功响应：

```json
{
  "code": 200,
  "data": {
    "service": {
      "user_id": 52,
      "nickname": "Hyper 客服",
      "avatar": "https://cdn.hypercn.cn/service.png",
      "signature": "在线客服"
    },
    "customer": {
      "user_id": 35,
      "nickname": "农子健",
      "avatar": "https://cdn.hypercn.cn/avatars/35.png",
      "signature": ""
    },
    "list": [
      {
        "id": 2060000000000000000,
        "sender_id": 35,
        "content": "退款什么时候到账？",
        "msg_type": 1,
        "ext": {},
        "time": 1786500000000,
        "is_self": false
      },
      {
        "id": 2060000000000000001,
        "sender_id": 52,
        "content": "已为您提交退款申请，请耐心等待处理。",
        "msg_type": 1,
        "ext": {},
        "time": 1786500030000,
        "is_self": true
      }
    ],
    "next_cursor": 1786500000000
  }
}
```

`is_self=true` 表示平台客服发送。前端不要用当前管理员 ID 判定消息左右侧。

若用户从未向客服账号发起过对话，返回：

```json
{
  "code": 404,
  "msg": "客服会话不存在，请等待用户先发起咨询"
}
```

## 3. 以平台客服身份回复

```http
POST /api/v1/admin/customer-service/sessions/35/messages
Content-Type: application/json
```

```json
{
  "msg_type": 1,
  "content": "已为您提交退款申请，请耐心等待处理。",
  "parent_msg_id": "0",
  "ext": {}
}
```

字段约定：

| 字段 | 必填 | 说明 |
|---|---:|---|
| `msg_type` | 是 | `1` 文本、`2` 图片、`3` 语音、`4` 视频、`5` 文件、`6` 位置、`7` 互动、`8` 卡片、`9` 活动卡片 |
| `content` | 是 | 文本内容或媒体 URL 等，最长 5000 字符 |
| `parent_msg_id` | 否 | 回复的消息 ID，默认 `0` |
| `ext` | 否 | 消息扩展字段 |

成功响应为现有 IM 消息对象：

```json
{
  "code": 200,
  "data": {
    "msg_id": "2060000000000000001",
    "sender_id": "52",
    "target_id": "35",
    "session_type": 1,
    "msg_type": 1,
    "content": "已为您提交退款申请，请耐心等待处理。",
    "timestamp": 1786500030000,
    "status": 0,
    "channel": "chat"
  }
}
```

消息先进入既有 `IM_CHAT_MSGS`，由 conn-server 落库、更新双方会话并尽力 Socket 推送。接口成功代表已进入消息队列；用户离线时仍能在下次进入小程序时通过会话历史查看。

## 4. 标记客服会话已读

```http
POST /api/v1/admin/customer-service/sessions/35/read
Content-Type: application/json
```

```json
{
  "read_time": 1786500030000
}
```

`read_time` 是当前已看到的最后消息时间（毫秒）。不传或传 `0` 时后端使用当前时间。

成功响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

后端仅在会话最后一条消息时间不晚于 `read_time` 时清零未读数，避免管理员正在查看时客户发来新消息却被误标已读。

## PC 页面接入建议

新增管理端菜单“客服工作台”：

1. 进入页面请求会话列表。
2. 选中会话后请求聊天记录，并立即调用已读接口。
3. 回复成功后把接口返回消息追加到当前列表；可用 `since=<最后消息时间>` 轮询补拉。
4. 前端展示“平台客服”身份，不展示具体管理员名称给客户端用户。
5. 若返回 `409` 配置缺失，展示“请先在系统配置中设置客服聊天用户 ID”。

## 权限与审计

- 新增权限码：`admin.customer_service`。
- 超级管理员 `*` 自动拥有权限。
- 非超级管理员需在角色 `permissions` 中显式加入该权限。
- 发送回复写入操作日志：`admin.customer_service.reply`。
- 标记已读写入操作日志：`admin.customer_service.read`。
- 登录态、私信范围和客服账号有效性均由后端二次校验，前端参数不能绕过。

## 与微信联动的后续方案

本期不做微信消息双向同步。后续可在本工作台稳定后增加：

1. 小程序订阅消息，仅提醒“客服已回复，请进入小程序查看”。
2. 微信原生客服作为独立入口。
3. 若要做真正双向同步，新增 `channel=wechat_kf`、微信回调验签、外部会话 ID 映射和幂等去重；不建议直接把现有 IM Topic 原样转发到微信。
