# 群组、成员与群聊接口文档

更新时间：2026-08-12

## 1. 范围与约定

本文覆盖小程序群聊所需的核心接口：建群、群资料、成员与权限管理、群消息和群会话。服务部署在网关后的完整路径统一为 `/api/v1/...`。

除非特别标注，所有接口均需：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

统一响应外层：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

前端应以响应体的 `code` 判断结果，不能只依赖 HTTP 状态码。

### ID 与角色

- `group_id`：群 ID，数值型。
- 群聊消息和群会话中，`peer_id` / `target_id` 均为 `group_id`。
- `session_type=2` 表示群聊；`session_type=1` 为私聊。
- `role=1` 群主，`role=2` 管理员，`role=3` 普通成员。
- 所有管理接口均由后端二次校验。前端的权限字段仅用于控制入口，不可替代后端鉴权。

## 2. 权限矩阵

| 操作 | 群主 | 管理员 | 普通成员 |
|---|:---:|:---:|:---:|
| 查看群资料、成员和历史消息 | 是 | 是 | 是 |
| 发送群消息 | 是 | 是 | 是，未被禁言时 |
| 邀请成员（直接入群） | 是 | 是 | 否 |
| 踢出普通成员 | 是 | 是 | 否 |
| 踢出管理员 | 是 | 否 | 否 |
| 单独禁言/解除普通成员 | 是 | 是 | 否 |
| 单独禁言/解除管理员 | 是 | 否 | 否 |
| 开启/关闭全员禁言 | 是 | 是 | 否 |
| 设置/撤销管理员 | 是 | 否 | 否 |
| 修改群名称、头像、简介 | 是 | 否 | 否 |
| 转让群主 | 是 | 否 | 否 |
| 解散群聊 | 是 | 否 | 否 |
| 退出群聊 | 退出即解散群 | 是 | 是 |

说明：

- 角色层级为 `群主 > 管理员 > 普通成员`。管理者只能操作低于自身角色的成员，不能操作自己。
- 开启全员禁言后，仅限制普通成员；群主和管理员仍可发送消息。
- 单独禁言优先级高于全员禁言。被单独禁言的管理员也不能发言。
- 群主调用“退出群聊”会解散群，不会留下无群主群组。前端群主应显示“解散群聊”，不要显示普通的“退出群聊”。

## 3. 群资料与成员总览

这是进入微信群资料页时的主接口，一次返回群资料、当前用户权限和所有有效成员。

```http
GET /api/v1/groupmember/list?group_id=8
```

兼容历史参数 `groupId`，新代码统一传 `group_id`。

### 成功响应

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "group": {
      "id": 8,
      "name": "周末骑行群",
      "avatar": "https://cdn.hypercn.cn/groups/8.png",
      "description": "每周末成都周边骑行",
      "owner_id": 1001,
      "member_count": 12,
      "max_members": 200,
      "is_mute_all": false,
      "created_at": "2026-08-01 10:00:00",
      "member_avatar_list": [
        "https://cdn.hypercn.cn/avatars/1001.png",
        "https://cdn.hypercn.cn/avatars/1002.png"
      ]
    },
    "current_user": {
      "user_id": 1001,
      "role": 1,
      "role_name": "群主",
      "user_card": "",
      "display_name": "小瑞",
      "is_owner": true,
      "is_admin": false,
      "is_muted": false,
      "can_send_message": true,
      "permissions": {
        "can_invite": true,
        "can_manage_members": true,
        "can_mute_members": true,
        "can_mute_all": true,
        "can_set_admin": true,
        "can_transfer_owner": true,
        "can_update_group_info": true,
        "can_dismiss_group": true,
        "can_quit": true
      }
    },
    "members": [
      {
        "user_id": 1001,
        "avatar": "https://cdn.hypercn.cn/avatars/1001.png",
        "nickname": "小瑞",
        "display_name": "小瑞",
        "gender": 3,
        "motto": "周末见",
        "role": 1,
        "role_name": "群主",
        "is_mute": 0,
        "is_muted": false,
        "can_send_message": true,
        "user_card": "",
        "join_time": "2026-08-01 10:00:00",
        "is_current_user": true,
        "can_kick": false,
        "can_mute": false,
        "can_set_admin": false,
        "can_transfer_owner": false
      },
      {
        "user_id": 1002,
        "avatar": "https://cdn.hypercn.cn/avatars/1002.png",
        "nickname": "阿琳",
        "display_name": "骑行领队",
        "gender": 2,
        "motto": "一路向前",
        "role": 3,
        "role_name": "群成员",
        "is_mute": 0,
        "is_muted": false,
        "can_send_message": true,
        "user_card": "骑行领队",
        "join_time": "2026-08-02 11:00:00",
        "is_current_user": false,
        "can_kick": true,
        "can_mute": true,
        "can_set_admin": true,
        "can_transfer_owner": true
      }
    ]
  }
}
```

### 字段说明

#### `data.group`

| 字段 | 说明 |
|---|---|
| `avatar` | 群头像；为空时，可用 `member_avatar_list` 生成九宫格头像。 |
| `member_count` | 以未退群成员实时计算的数量。 |
| `max_members` | 当前创建群默认值为 `200`。 |
| `is_mute_all` | 是否开启全员禁言。 |
| `member_avatar_list` | 最多 9 个有效成员头像，顺序为群主、管理员、普通成员。 |

#### `data.current_user`

`permissions` 为群资料页操作入口的唯一依据：

| 字段 | 前端入口 |
|---|---|
| `can_invite` | 邀请成员 |
| `can_manage_members` | 成员管理菜单 |
| `can_mute_members` | 成员禁言/解除禁言 |
| `can_mute_all` | 全员禁言开关 |
| `can_set_admin` | 设置/撤销管理员 |
| `can_transfer_owner` | 转让群主 |
| `can_update_group_info` | 编辑群名称、头像、简介 |
| `can_dismiss_group` | 解散群聊 |
| `can_quit` | 退出群聊；群主点击后实际解散群 |

#### `data.members[]`

| 字段 | 说明 |
|---|---|
| `display_name` | 已按“群名片优先、昵称其次、用户 ID 兜底”计算，列表直接展示。 |
| `motto` | 用户个性签名。 |
| `role` / `role_name` | 成员角色及中文展示名。 |
| `is_muted` | 是否处于单独禁言。 |
| `can_send_message` | 已综合单独禁言、全员禁言和角色计算。 |
| `join_time` | 入群时间，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `can_kick` / `can_mute` | 当前用户是否可对该成员踢出或禁言。 |
| `can_set_admin` / `can_transfer_owner` | 当前用户是否可对该成员设置管理员或转让群主。 |

成员顺序已经是群主、管理员、普通成员，前端不要重新按昵称排序；成员操作菜单只按当前成员项的 `can_*` 字段显示。

## 4. 建群与群资料维护

### 4.1 创建群聊

```http
POST /api/v1/group/create
```

请求：

```json
{
  "name": "周末骑行群",
  "avatar": "https://cdn.hypercn.cn/groups/8.png",
  "description": "每周末成都周边骑行"
}
```

| 字段 | 必填 | 规则 |
|---|:---:|---|
| `name` | 是 | 1-100 个字符 |
| `avatar` | 否 | 群头像 URL |
| `description` | 否 | 最长 500 个字符 |

成功响应：

```json
{
  "code": 200,
  "data": {
    "id": 8,
    "name": "周末骑行群",
    "avatar": "https://cdn.hypercn.cn/groups/8.png",
    "owner_id": 1001,
    "member_count": 1,
    "created_at": "2026-08-12 18:00:00",
    "session_id": "123"
  }
}
```

创建人自动成为群主，群最大人数为 `200`。`session_id` 仅用于会话定位；进入群聊、读消息和管理成员统一使用返回的 `id` 作为 `group_id`。

### 4.2 修改群名称

```http
POST /api/v1/group/update-name
```

```json
{
  "group_id": 8,
  "name": "成都周末骑行群"
}
```

仅群主可调用，`name` 长度为 1-20 个字符。成功后返回 `data: "群名称已更新"`。

### 4.3 修改群头像

```http
POST /api/v1/group/update-avatar
```

```json
{
  "group_id": 8,
  "avatar": "https://cdn.hypercn.cn/groups/8-new.png"
}
```

仅群主可调用。成功后返回 `data: "群头像已更新"`。

### 4.4 修改群简介

```http
POST /api/v1/group/update-description
```

```json
{
  "group_id": 8,
  "description": "每周末成都周边骑行，欢迎新朋友"
}
```

仅群主可调用，`description` 长度最长 100 个字符。成功后返回 `data: "群描述已更新"`。

### 4.5 解散群聊

```http
POST /api/v1/group/dismiss
```

```json
{
  "group_id": 8
}
```

仅群主可调用。解散后全体成员被移出、群会话从会话列表移除，成功返回 `data: "群解散成功"`。该动作不可恢复，前端必须二次确认。

## 5. 成员管理

### 5.1 按手机号查找可邀请用户

群主或管理员在“邀请成员”页输入完整 11 位手机号后，先调用本接口取得可选用户；不要让用户输入或猜测 `user_id`。

```http
POST /api/v1/groupmember/invite-candidate
```

请求：

```json
{
  "group_id": 8,
  "mobile": "13800138000"
}
```

| 字段 | 必填 | 说明 |
|---|:---:|---|
| `group_id` | 是 | 当前群 ID。 |
| `mobile` | 是 | 中国大陆 11 位手机号，仅支持精确查询。 |

仅群主和管理员可以查询；普通成员、非成员和已退出成员无法调用。接口仅返回脱敏手机号，不返回完整手机号。

找到可邀请用户：

```json
{
  "code": 200,
  "data": {
    "found": true,
    "candidate": {
      "user_id": 1003,
      "mobile_masked": "138****8000",
      "nickname": "阿琳",
      "avatar": "https://cdn.hypercn.cn/avatars/1003.png",
      "motto": "一路向前",
      "membership_status": "not_member",
      "can_invite": true,
      "invite_disabled_reason": ""
    }
  }
}
```

未查询到可邀请用户时为正常空结果，不按错误处理：

```json
{
  "code": 200,
  "data": {
    "found": false,
    "candidate": null
  }
}
```

`membership_status` 取值：

| 值 | 含义 | `can_invite` |
|---|---|:---:|
| `not_member` | 从未加入该群 | `true` |
| `left` | 曾加入但已退群或被踢出 | `true`，邀请后恢复入群 |
| `active` | 当前已在群内 | `false` |

当群人数已达上限时，候选用户仍会返回，但 `can_invite=false` 且 `invite_disabled_reason` 为“群人数已达上限”。前端直接展示该提示，不要提交邀请请求。

手机号查找仅用于邀请确认，不能据此展示完整通讯录或做模糊手机号搜索。手机号通过 JSON 请求体提交，避免出现在 URL 中。

### 5.2 确认邀请成员

```http
POST /api/v1/groupmember/invite
```

```json
{
  "group_id": 8,
  "invited_user_ids": [1002, 1003]
}
```

群主和管理员可调用。`invited_user_ids` 必须来自上一接口的 `candidate.user_id`，前端可一次选择多个搜索结果再提交。后端仍会校验用户存在、账号正常、不能邀请自己且不会重复处理同一 ID。

该接口为**直接加入群聊**，当前没有“待确认邀请”或“申请入群”状态。已在群内的用户会记入失败列表；已退出的旧成员会恢复为正常成员。

```json
{
  "code": 200,
  "data": {
    "success_count": 1,
    "failed_count": 1,
    "failed_user_ids": [1002]
  }
}
```

成功加入的成员会自动获得该群会话。

### 5.3 踢出成员

```http
POST /api/v1/groupmember/kick
```

```json
{
  "group_id": 8,
  "kicked_user_id": 1003
}
```

群主可踢管理员和普通成员；管理员只能踢普通成员。不能踢自己、不能踢同级或更高角色。成功响应：

```json
{
  "code": 200,
  "data": { "success": true }
}
```

被踢成员的群会话及未读数会清理。

### 5.4 单独禁言或解除禁言

```http
POST /api/v1/groupmember/mute
```

```json
{
  "group_id": 8,
  "target_user_id": 1003,
  "mute": true
}
```

- `mute=true`：禁言。
- `mute=false`：解除禁言。
- 群主可操作管理员和普通成员；管理员仅可操作普通成员；群主不可被禁言。

成功响应：`data: "ok"`。

### 5.5 全员禁言

```http
POST /api/v1/groupmember/mute-all
```

```json
{
  "group_id": 8,
  "mute": true
}
```

群主和管理员可调用。`mute=true` 开启，`mute=false` 关闭。全员禁言不限制群主和管理员发言。

```json
{
  "code": 200,
  "data": {
    "is_mute_all": true
  }
}
```

### 5.6 设置或撤销管理员

```http
POST /api/v1/groupmember/admin
```

```json
{
  "group_id": 8,
  "target_user_id": 1003,
  "admin": true
}
```

仅群主可调用。`admin=true` 设为管理员，`admin=false` 撤销管理员；不能操作群主。成功响应：`data: "ok"`。

### 5.7 转让群主

```http
POST /api/v1/groupmember/transfer-owner
```

```json
{
  "group_id": 8,
  "new_owner_id": 1003
}
```

仅群主可调用。新群主必须是当前未退群成员，且不能转让给自己。成功后旧群主自动降级为普通成员。

```json
{
  "code": 200,
  "data": { "success": true }
}
```

### 5.8 退出群聊

```http
POST /api/v1/groupmember/quit
```

```json
{
  "group_id": 8
}
```

普通成员和管理员退出后，会从群成员、会话列表和未读统计中移除：

```json
{
  "code": 200,
  "data": { "disbanded": false }
}
```

群主调用时会直接解散群：

```json
{
  "code": 200,
  "data": { "disbanded": true }
}
```

## 6. 群消息

### 6.1 发送群消息

```http
POST /api/v1/message/send
```

文本消息请求：

```json
{
  "target_id": 8,
  "session_type": 2,
  "msg_type": 1,
  "content": "周六早上九点集合",
  "parent_msg_id": "0",
  "ext": {}
}
```

| 字段 | 说明 |
|---|---|
| `target_id` | 必填，群 ID。 |
| `session_type` | 必填，群聊固定为 `2`。 |
| `msg_type` | `1` 文本、`2` 图片、`3` 语音、`4` 视频、`5` 文件、`6` 位置、`7` 互动、`8` 卡片、`9` 活动卡片。 |
| `content` | 消息正文或资源地址，由消息类型决定。 |
| `parent_msg_id` | 可选，回复的父消息 ID；大整数以字符串传递。 |
| `ext` | 可选扩展对象；转发帖子/活动卡片时由后端补充可信卡片信息。 |

`sender_id`、`msg_id`、`timestamp`、`status`、`session_id`、`session_hash`、`channel` 均由服务端生成或覆盖，前端无需也不能信任自行传入的值。

只有当前未退群成员能发送消息；单独禁言成员不能发送；全员禁言期间普通成员不能发送，群主和管理员不受影响。

### 6.2 拉取群消息历史

```http
GET /api/v1/message/list?peer_id=8&session_type=2&limit=20&cursor=0&since=0
```

| 参数 | 必填 | 说明 |
|---|:---:|---|
| `peer_id` | 是 | 群 ID。 |
| `session_type` | 是 | 群聊固定为 `2`。 |
| `limit` | 否 | 默认 20，最大 100。 |
| `cursor` | 否 | 向上翻历史时传已加载列表中最早消息的 `time`（毫秒）。 |
| `since` | 否 | 拉取新消息时传最后一条已知消息的 `time`（毫秒）；传入后忽略 `cursor`。 |

该接口已强制校验群有效且当前用户是未退群成员；退群用户、被踢用户和非成员不能读取群历史。

```json
{
  "code": 200,
  "data": {
    "self_avatar": "https://cdn.hypercn.cn/avatars/1001.png",
    "list": [
      {
        "id": 206000000000001,
        "sender_id": 1002,
        "nickname": "阿琳",
        "avatar": "https://cdn.hypercn.cn/avatars/1002.png",
        "content": "周六早上九点集合",
        "msg_type": 1,
        "ext": {},
        "time": 1786528800000,
        "is_self": false
      }
    ],
    "next_cursor": 1786528800000,
    "unread_total": 3
  }
}
```

`list` 已按时间正序返回。群聊响应中的 `avatar`、`nickname` 为空字符串；消息发送者信息应读取每个 `list[]` 项的 `avatar`、`nickname`。

## 7. 群会话

### 7.1 获取会话列表

```http
GET /api/v1/session/
```

群会话项满足 `session_type=2`，其 `peer_id` 即群 ID：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "session_type": 2,
        "peer_id": 8,
        "peer_name": "周末骑行群",
        "peer_avatar": "https://cdn.hypercn.cn/groups/8.png",
        "last_msg": "周六早上九点集合",
        "last_msg_time": 1786528800000,
        "unread": 3,
        "is_top": 0,
        "is_mute": 0
      }
    ]
  }
}
```

### 7.2 置顶和会话免打扰

```http
POST /api/v1/session/setting
```

```json
{
  "session_type": 2,
  "peer_id": 8,
  "is_top": 1,
  "is_mute": 0
}
```

`is_top` 和 `is_mute` 都必须传 `0` 或 `1`。这里的 `is_mute` 是**当前用户的会话通知免打扰**，不会禁言，也不影响其他成员；不要与群成员的 `is_muted` 或群资料的 `is_mute_all` 混用。

成功响应：`data: "ok"`。

### 7.3 清除群会话未读数

```http
POST /api/v1/session/clear-unread
```

```json
{
  "session_type": 2,
  "peer_id": 8,
  "read_time": 1786528800000
}
```

`read_time` 可选，单位为毫秒。建议前端进入群聊并渲染到最后消息后传最后消息时间，后端只会在没有更晚消息时清零，避免清掉刚到达的新未读。

## 8. 前端接入顺序与刷新规则

1. 创建群后，用响应的 `id` 请求 `GET /groupmember/list`，再跳转群聊页。
2. 会话列表中找 `session_type=2` 的项，使用其 `peer_id` 作为群 ID；不要把 `session_id` 当作群 ID。
3. 群主或管理员邀请成员时，输入完整手机号并调用 `POST /groupmember/invite-candidate`；当 `found=true` 且 `candidate.can_invite=true` 时，才允许选中该用户。
4. 将所有选中候选人的 `candidate.user_id` 放入 `POST /groupmember/invite` 的 `invited_user_ids`；成功后重新请求 `GET /groupmember/list`。
5. 踢人、禁言、全员禁言、设管理员、转让群主、修改群资料成功后，重新拉取 `GET /groupmember/list`。
6. 退出或解散成功后回到会话列表，并重新请求 `GET /session/`；不要继续访问已退出或已解散的群。
7. 发送前可用 `current_user.can_send_message` 禁用输入框，但实际结果必须以发送接口返回为准。

## 9. 错误码与当前缺口

常见业务 `code`：

| code | 场景 | 前端处理 |
|---:|---|---|
| `400` | 参数错误、重复操作、成员不在群内、无可操作权限 | 展示 `msg`，保留当前页。 |
| `401` | 未登录或 Token 失效 | 重新登录。 |
| `403` | 查询群资料或群消息时已退群/不是成员 | 退出当前群页面并刷新会话列表。 |
| `404` | 群不存在或已解散 | 提示群已不可用并回退。 |
| `500` | 数据库、消息队列或其他服务端异常 | 显示通用重试提示。 |

当前已明确**未提供**以下接口，前端不要预留成可用功能：

- 用户主动申请入群、群主审批入群。
- 修改本人群名片 `user_card`。
- 成员邀请链接、二维码入群。
- 群公告的独立发布/历史接口；当前 `description` 仅为群简介。
- 群消息撤回、删除、置顶公告等管理接口。
