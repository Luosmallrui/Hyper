# 消息接口文档（IM）

**接口说明**:：本文档中除非特别注明，均需要认证。

**备注**：
```
当 access token 即将过期时，服务端可能在响应头返回 X-New-Access-Token，客户端可自行替换使用。
（未登录会返回 401）

```

**接口列表**：
```
1）发送消息
POST /v1/message/send（需要认证）
说明：发送消息（单聊/群聊）

2）拉取消息列表
GET /v1/message/list（需要认证）
说明：拉取指定会话的历史消息列表（相当于加载聊天记录）

3）清除会话未读数
POST /v1/message/clear-unread（需要认证）
说明：清除指定会话的未读数（服务端实现为重置未读，相当于标记这个会话已读）

4)获取会话列表
GET /v1/session/list（需要认证）
说明：获取当前用户的会话列表（单聊/群聊混合），用于会话列表页展示（相当于QQ 主页“消息列表刷新”）

5) 会话设置（置顶/免打扰）
POST /v1/session/setting（需要认证）
说明：更新当前用户某个会话的设置项（如置顶/免打扰）

6) 创建群聊（建群）
POST /group/create（需要认证）
说明：创建一个群聊，并自动把创建者加入群成员（创建者为群主）

7) 邀请成员入群（拉人）
POST /groupmember/invite（需要认证）
说明：邀请多个用户加入指定群聊（当前实现为“直接入群”，无需审批）

8) 踢出群成员（踢人）
POST /groupmember/kick（需要认证）
说明：将指定成员踢出群聊（服务端实现为把该成员标记为退群 is_quit=1，并扣减群成员数

9) 获取群成员列表
GET /groupmember/list（需要认证）
说明：获取指定群聊的成员列表（仅返回未退群成员）

) 建立 WebSocket 连接（IM）(未完成)
WebSocket /im/wss（需要认证）
说明：建立 IM WebSocket 长连接（用于实时消息推送/心跳/ACK）。
握手方式：HTTP GET /im/wss + Upgrade: websocket（成功返回 101 Switching Protocols）

```

## 1) 发送消息
```
POST /v1/message/send（需要认证）
说明：发送一条消息（支持单聊/群聊）
sender_id 会由服务端根据 token 自动填写，前端传不传都无效/会被覆盖

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)
请求示例：
```json
{
  "target_id": "10",
  "session_type": 1,
  "msg_type": 1,
  "content": "哇咔咔",
  "parent_msg_id": "0",
  "ext": {
    "device": "iPhone 15"
  },
  "channel": "chat"
}
```
### 参数说明

| 字段 | 类型 | 必填 | 说明 |
|----|----|----|----|
| target_id | string | 是  | 接收者ID（单聊对方 user_id）或群ID（群聊 group_id） |
| session_type | int | 是  | 会话类型：1=单聊，2=群聊 |
| msg_type | int | 是  | 消息类型：1文本，2图片，3语音，4视频，5文件，6位置，7互动，8卡片 |
| content	 | string | 否  | 消息内容（文本/URL/描述等，依业务而定） |
| parent_msg_id | string | 否  | 回复消息ID（不回复可传 0 或不传） |
| timestamp | int | 否  | 时间戳（毫秒）服务端统一生成/覆盖） |
| ext	 | object | 否  | 扩展字段（JSON 对象） |
| channel | string | 否  | 渠道：chat/system/notification/control/heartbeat（当前发送接口会统一写成chat） |
| sender_id | string | 否  |  发送者ID（前端传无效，服务端从 token 覆盖）  |

## 响应结果

**统一响应格式**:
```
成功响应：
{"code":200,"msg":"ok","data":{...}}

失败响应（业务错误/参数错误等）：
{"code":xxx,"msg":"错误信息"}

```
### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "msg_id": "2012463600169390080",
    "sender_id": "9",
    "target_id": "10",
    "session_type": 1,
    "session_hash": -8567956223700959152,
    "session_id": "9_10",
    "msg_type": 1,
    "content": "哇咔咔",
    "parent_msg_id": "0",
    "timestamp": 1768643686703,
    "status": 0,
    "ext": {
      "device": "iPhone 15"
    },
    "channel": "chat"
  }
}
```

### 失败响应

```json
{
  "code": 401,
  "message":"未登录"
}
```

```json
{
  "code": 500,
  "message": "xxx(参数解析失败/服务端错误)"
}
```

## 状态码说明

| 状态码 | 说明 | 触发场景 | 
|-----|----|------|
| 200 | 成功 |  发送成功进入 MQ 流程   |
| 401 | 未登录/鉴权失败 | 无法从 context 获取 user_id（token 无效/缺失） |
| 500 | 服务器内部错误 | JSON 解析失败或 service 返回 error |

## 数据库字段映射
**注意：本接口本身只负责生成消息并投递 MQ，不直接落库**

**单聊：im_single_messages**

| 数据库字段 | 说明 |
|-----|------|
| id  | 消息ID（雪花ID，主键） |
| session_hash | 会话 hash（用于索引/分页）|
| session_id | 会话ID（如 8_10） |
| sender_id | 发送者 uid |
| target_id | 接收者 uid |
| msg_type | 消息类型 |
| content | 消息内容 |
| parent_msg_id | 回复消息ID|
| status | 落库写 成功(1) |
| ext | 扩展字段（JSON） |
| created_at| 毫秒时间戳|


**群聊：im_group_messages**

| 数据库字段 | 说明 |
|-----|------|
| id  | 消息ID（雪花ID，主键） |
| session_hash | 群会话 hash（用于索引/分页）|
| session_id | 群会话ID（g_<groupId>） |
| sender_id | 发送者 uid |
| target_id | 群ID（groupId） |
| msg_type | 消息类型 |
| content | 消息内容 |
| parent_msg_id | 回复消息ID|
| status | 落库写 成功(1) |
| ext | 扩展字段（JSON） |
| created_at| 毫秒时间戳|


## 2) 拉取消息列表
```
GET /v1/message/list（需要认证）
说明：拉取某个会话的历史消息列表（单聊/群聊通用）（相当于加载聊天记录）
支持 cursor/since 分页（具体分页含义以服务端实现为准，推荐用服务端返回的 next_cursor 做翻页）

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)

请求示例：
`GET v1/message/list?peer_id=5&session_type=2&cursor=0&limit=20`

### 查询参数

| 参数 | 类型 | 必填 | 默认 | 说明 |
|----|----|----|----|----|
| peer_id | uint64 | 是  | -  | 对端ID：单聊为对方 user_id，群聊为 group_id |
| session_type | int | 否  | 1  | 1=单聊，2=群聊（只能是 1 或 2） |
| cursor | int64 | 否  | 0  | 游标（翻页拉旧消息） |
| since	 | int64 | 否  | 0  | 起始时间（可选） |
| limit | int | 否  | 20 | 拉取条数（service 侧会限制最大 100) |


### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "avatar": "",
    "list": [
      {
        "id": 2012079966677635072,
        "sender_id": 8,
        "content": "群聊历史测试：第一条",
        "msg_type": 1,
        "ext": {},
        "time": 1768552221351,
        "is_self": false
      }
    ],
    "next_cursor": 1768552221351,
    "nickname": "",
    "self_avatar": "https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0"
  }
}
```

## 响应参数说明

| 字段 | 类型 | 说明 | 
|----|----|----|
| avatar | string | 对方头像（仅单聊；群聊返回空字符串） |
| nickname | string | 对方昵称（仅单聊；群聊返回空字符串） |
| self_avatar | string | 自己头像 |
| list   | array |     消息列表 |
| next_cursor | int64 |  下一页游标（当前实现为“本次返回中最老一条消息的 time”） |

list 元素结构

| 字段 | 类型  | 说明 | 
|----|-----|----|
| id | uint64 | 消息ID |
| sender_id | uint64 | 发送者ID |
| content | string | 消息内容 |
| msg_type | int | 消息类型 |
| ext | object | 扩展字段  |
|  time   | int64 | 消息时间（用于排序/翻页） |
| is_self  | bool |  是否自己发送  |



## 3) 清除会话未读数
```
POST /v1/message/clear-unread（需要认证）
说明：清除指定会话的未读数（服务端实现为 Reset/重置，相当于标记这个会话已读）
常见使用场景：进入会话页后调用一次

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)

请求示例：
```json
{
  "session_type": 2,
  "peer_id": 5
}
```
### 请求参数

| 参数 | 类型   | 必填 | 说明  |
|----|------|----|-----|
| session_type | int  | 是  | 1=单聊，2=群聊（只允许 1 或 2) |
| peer_id | uint | 是  | 对端ID（单聊为对方 user_id，群聊为 group_id） |

### 成功响应
成功示例：

```json
{
  "code": 200,
  "msg": "ok",
  "data": "ok"
}
```


## 4) 获取会话列表
```
GET /v1/session/list（需要认证）
说明：获取当前用户的会话列表（单聊/群聊混合），用于会话列表页展示（相当于QQ 主页“消息列表刷新”）

```

**请求头**:
```
Authorization: Bearer <token>

```

## 请求参数
无

### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "list": [
      {
        "session_type": 1,
        "peer_id": 9,
        "last_msg": "你是谁",
        "last_msg_time": 1767884209419,
        "unread": 7,
        "is_top": 0,
        "is_mute": 0,
        "peer_avatar": "https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0",
        "peer_name": "泥嚎"
      },
      {
        "session_type": 2,
        "peer_id": 5,
        "last_msg": "拉回后测试：10应该收到",
        "last_msg_time": 1768579935905,
        "unread": 0,
        "is_top": 0,
        "is_mute": 0,
        "peer_avatar": "https://cdn.hypercn.cn/avatars/05/5/876d0321.jpg",
        "peer_name": "鸟鸟"
      }
    ]
  }
}
```

## 响应参数说明
data 字段

| 字段 | 类型 | 说明 | 
|----|----|----|
| list   | array |会话列表（SessionDTO 数组） |


list 元素结构

| 字段 | 类型 | 说明 | 
|----|----|----|
| session_type| int | 会话类型：1=单聊，2=群聊 |
| peer_id | uint64 | 对端ID：单聊为对方 user_id；群聊为 group_id |
| last_msg | string | 最后一条消息摘要 |
| last_msg_time | int64 | 最后一条消息时间戳(毫秒) |
| unread | int | 未读数（进入会话后可通过 clear-unread 清零） |
|  is_top   | int | 是否置顶：0否/1是 |
| is_mute   | int | 是否免打扰：0否/1是 |
| peer_avatar | string | 对端头像：单聊为对方用户头像；群聊为群头像（或群主/群资料头像，依实现） |
| peer_name | string | 对端名称：单聊为对方昵称；群聊为群名 |



## 5) 会话设置（置顶/免打扰）
```
POST /v1/session/setting（需要认证）
说明：更新当前用户某个会话的设置项（如置顶/免打扰）
使用场景：在会话列表页对某个会话进行“置顶/取消置顶”“免打扰/取消免打扰”等操作

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)

请求示例：
```json
{
  "session_type": 1,
  "peer_id": 10,
  "is_top": 1,
  "is_mute": 0
}

```
### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|----|----|---|----|
| session_type | int | 是 | 1=单聊，2=群聊（只允许 1 或 2) |
| peer_id | uint | 是 | 对端ID（单聊为对方 user_id，群聊为 group_id） |
| is_top | int | 是 | 是否置顶：0=否，1=是 |
| is_mute | int | 是 | 是否免打扰：0=否，1=是 |


### 成功响应
成功示例：

```json
{
  "code": 200,
  "msg": "ok",
  "data": "ok"
}
```
### 失败响应

```json
{
  "code": 400,
  "msg": "参数错误"
}
```
```json
{
  "code": 401,
  "msg": "未登录"
}
```
```json
{
  "code": 500,
  "msg": "xxx(服务端错误)"
}
```
说明：该接口仅更新设置项，不返回会话详情；前端通常更新本地状态或重新拉取会话列表。


## 6) 创建群聊（建群）
```
POST /group/create（需要认证）
说明：创建一个群聊，并自动把创建者加入群成员（创建者为群主）
```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)

请求示例：
```json
{
  "name": "这不对吧",
  "avatar": "https://example.com/group/avatar.png",
  "description": "come on！"
}
```
### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|----|----|---|----|
| name | string | 是 | 群名称（1~100 字符） |
| avatar | string | 否 | 群头像 URL（不传则默认为空字符串） |
| description | string | 否 | 群描述（最大 500 字符） |

### 成功响应
成功示例：
```json
{
  "code": 200,
  "data": {
    "id": 6,
    "name": "这不对吧",
    "avatar": "https://example.com/group/avatar.png",
    "owner_id": 8,
    "member_count": 1,
    "created_at": "2026-01-18 15:28:30"
  },
  "message": "创建群成功"
}
```

## 响应参数说明
data 字段

| 字段 | 类型 | 说明 | 
|----|----|----|
| id| int | 群名称 |
| name | string | 群头像 |
| avatar | string | 群主用户ID（由服务端从 token 解析得到） |
| owner_id | int | 当前群成员数量（创建时固定为 1） |
| member_count | int | 未读数（进入会话后可通过 clear-unread 清零） |
|  created_at   | string | 创建时间（格式：YYYY-MM-DD HH:mm:ss） |

## 数据库字段映射
**groups**

| 数据库字段 | 类型 | 说明 |
|-----|----|------|
| name  | varchar(100) | 群名称 |
| avatar | varchar(255) | 群头像（默认空字符串）|
| description | varchar(500) | 群描述（默认空字符串） |
| owner_id | int unsigned | 群主用户ID |
| member_count | int unsigned |初始成员数 |
| max_members | int unsigned | 最大成员数 |
| created_at | datetime |创建时间 |
| updated_at | datetime | 更新时间|

## 7) 邀请成员入群（拉人）
```
POST /groupmember/invite（需要认证）
说明：邀请多个用户加入指定群聊（当前实现为“直接入群”，无需审批）

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)
请求示例：
```json
{
  "group_id": 6,
  "user_ids": [9, 10]
}
```
### 参数说明

| 字段 | 类型 | 必填 | 说明 |
|----|----|----|----|
| group_id | int | 是  | 群ID（groups.id） |
| user_ids | array[int] | 是  | 被邀请用户ID列表 |
说明：邀请者用户ID由服务端从 token 解析得到，前端无需传。


### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "invited_members": {
      "success_count": 2,
      "failed_count": 0,
      "failed_user_ids": []
    }
  }
}
```

## 响应参数说明
data 字段

| 字段 | 类型 | 说明 | 
|----|----|----|
| success_count| int | 邀请成功的人数 |
| failed_count | int | 邀请失败的人数 |
| failed_user_ids | array[int] | 邀请失败的用户ID列表（如：已在群内、邀请自己、DB写入失败等） |


## 8) 踢出群成员（踢人）
```
POST /groupmember/kick（需要认证）
说明：将指定成员踢出群聊（服务端实现为把该成员标记为退群 is_quit=1，并扣减群成员数）

```

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json

```

## 请求参数

### 请求体 (JSON)
请求示例：
```json
{
  "group_id": 6,
  "kicked_user_id": 10
}
```
### 参数说明

| 字段 | 类型 | 必填 | 说明 |
|----|----|----|----|
| group_id | int | 是  | 群ID（groups.id） |
| kicked_user_id | int | 是  | 被踢出的用户ID |
说明：操作者用户ID由服务端从 token 解析得到，前端无需传。

### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": "踢出成员成功"
}
```

## 9) 获取群成员列表
```
GET /groupmember/list（需要认证）
说明：获取指定群聊的成员列表。
- 仅返回未退群成员（group_member.is_quit = 0）
- 默认按成员角色升序排列：群主(1) -> 管理员(2) -> 普通成员(3)

```

**请求头**:
```
Authorization: Bearer <token>

```

## 请求参数

### 请求体 (JSON)

请求示例：
`GET /groupmember/list?group_id=6`

### 查询参数

| 参数 | 类型 | 必填 | 默认 | 说明 |
|----|----|----|----|----|
| group_id | int | 是  | -  | 群ID（groups.id） |

### 成功响应
成功示例：
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "members": [
      {
        "user_id": 8,
        "avatar": "https://cdn.hypercn.cn/avatars/08/8/5f349c00.jpeg",
        "nickname": "帜羲",
        "gender": 0,
        "motto": "",
        "role": 1,
        "is_mute": 0,
        "user_card": ""
      },
      {
        "user_id": 9,
        "avatar": "https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0",
        "nickname": "泥嚎",
        "gender": 0,
        "motto": "",
        "role": 3,
        "is_mute": 0,
        "user_card": ""
      }
    ]
  }
}
```


## 响应参数说明
data 字段

| 字段 | 类型 | 说明 | 
|----|----|----|
| members | array | 成员列表 |

members 元素结构

| 字段 | 类型 | 说明 | 
|----|----|----|
| user_id | int | 	成员用户ID |
| avatar | string | 用户头像 |
| nickname | string | 用户昵称 |
| gender | int | 用户性别（1:男 ;2:女;3:未知） |
| motto | string | 个性签名 |
| role | int | 成员角色：1群主/2管理员/3普通成员 |
| is_mute |   int  |   是否禁言：0否 / 1是 |
| user_card | string  |  群名片 |
说明：gender不知道为什么在数据库中所有人都是0，明明默认值是3

## 状态码说明

| 状态码 | 说明 | 触发场景 | 
|-----|----|------|
| 200 | 成功 |  返回成员列表 |
| 400 | 参数错误 | group_id 缺失或非法 |
| 401 | 未登录/鉴权失败 | token 缺失/无效 |
| 403 |  无权限 | 当前用户不在该群内（或已退群）|
| 500 |  服务器内部错误 | DB 查询失败 |


## ) 建立 WebSocket 连接（IM）（未完成）
```
WebSocket /im/wss（需要认证）
说明：建立 IM WebSocket 长连接（用于实时消息推送/心跳/ACK）。
握手方式：HTTP GET /im/wss + Upgrade: websocket（成功返回 101 Switching Protocols）

```

**连接地址**：
```
测试环境：ws://<host>/im/wss
生产环境：wss://<host>/im/wss
```

**请求头**:
```
Authorization: Bearer <token>

```
说明：Upgrade/Connection/Sec-WebSocket-* 等握手头一般由 WebSocket 客户端自动带上，你只需要保证 Authorization 存在即可。

## 请求参数
无

### 成功响应

```
成功响应（握手成功）
HTTP 状态码：101 Switching Protocols

```
注意：WebSocket 连接建立后，不会返回标准的 JSON 响应体（例如 {"code":200,...}），而是进入 WebSocket 双向消息通道。

### 失败响应

```
401 Unauthorized：token 缺失/无效/过期
403 Forbidden：无权限（如有权限控制）
404 Not Found：路由不存在/网关未转发 WebSocket
500 Internal Server Error：服务端内部错误
```

