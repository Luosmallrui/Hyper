# 笔记创建接口文档

## 接口信息

**接口路径**: `POST /v1/note/create`

**接口说明**: 创建新的笔记（需要用户认证）

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

## 请求参数

### 请求体 (JSON)

```json
{
  "title": "我的第一篇笔记",
  "content": "这是笔记的内容，支持长文本...",
  "topic_ids": [1001, 1002],
  "location": {
    "lat": 39.9042,
    "lng": 116.4074,
    "name": "北京天安门"
  },
  "media_data": [
    {
      "url": "https://example.com/image1.jpg",
      "thumbnail_url": "https://example.com/thumb1.jpg",
      "width": 1920,
      "height": 1080,
      "duration": 0
    }
  ],
  "type": 1,
  "visible_conf": 1
}
```

### 参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 笔记标题，最大长度 100 字符 |
| content | string | 否 | 笔记正文内容 |
| topic_ids | array | 否 | 话题 ID 列表 |
| location | object | 否 | 地理位置信息 |
| location.lat | float | 否 | 纬度 |
| location.lng | float | 否 | 经度 |
| location.name | string | 否 | 地点名称 |
| media_data | array | 否 | 媒体资源列表 |
| media_data[].url | string | 否 | 图片/视频 URL |
| media_data[].thumbnail_url | string | 否 | 缩略图 URL |
| media_data[].width | int | 否 | 宽度 |
| media_data[].height | int | 否 | 高度 |
| media_data[].duration | int | 否 | 视频时长（秒） |
| type | int | 是 | 笔记类型：1-图文, 2-视频 |
| visible_conf | int | 否 | 可见性：1-公开, 2-粉丝可见, 3-自己可见（默认 1） |

## 响应结果

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "note_id": 1234567890123456789
  }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "参数格式错误: title is required"
}
```

```json
{
  "code": 401,
  "message": "未授权"
}
```

```json
{
  "code": 500,
  "message": "创建笔记失败: database error"
}
```

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（token 无效或未提供） |
| 500 | 服务器内部错误 |

## 数据库字段映射

创建后的笔记会存储到 `note` 表中：

| 数据库字段 | 说明 | 默认值 |
|-----------|------|--------|
| id | 笔记 ID（雪花算法生成） | 自动生成 |
| user_id | 用户 ID | 从 token 中获取 |
| title | 标题 | - |
| content | 内容 | - |
| topic_ids | 话题 ID 列表（JSON） | - |
| location | 地理位置（JSON） | - |
| media_data | 媒体资源（JSON） | - |
| type | 类型 | - |
| status | 状态 | 0（审核中） |
| visible_conf | 可见性 | 1（公开） |
| created_at | 创建时间 | 当前时间 |
| updated_at | 更新时间 | 当前时间 |

## 笔记状态说明

| 状态值 | 说明 |
|-------|------|
| 0 | 审核中 |
| 1 | 公开 |
| 2 | 私密 |
| 3 | 违规 |

## 使用示例

### 使用 curl

```bash
curl -X POST http://localhost:8080/v1/note/create \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "今天的美食",
    "content": "今天去了一家很棒的餐厅",
    "topic_ids": [1001],
    "location": {
      "lat": 39.9042,
      "lng": 116.4074,
      "name": "北京"
    },
    "media_data": [
      {
        "url": "https://example.com/food.jpg",
        "thumbnail_url": "https://example.com/food_thumb.jpg",
        "width": 1920,
        "height": 1080
      }
    ],
    "type": 1,
    "visible_conf": 1
  }'
```

### 使用 JavaScript (fetch)

```javascript
const createNote = async () => {
  const response = await fetch('http://localhost:8080/v1/note/create', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer your_token_here',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      title: '今天的美食',
      content: '今天去了一家很棒的餐厅',
      topic_ids: [1001],
      location: {
        lat: 39.9042,
        lng: 116.4074,
        name: '北京'
      },
      media_data: [
        {
          url: 'https://example.com/food.jpg',
          thumbnail_url: 'https://example.com/food_thumb.jpg',
          width: 1920,
          height: 1080
        }
      ],
      type: 1,
      visible_conf: 1
    })
  });
  
  const data = await response.json();
  return data.data.note_id;
};
```

## 注意事项

1. 必须先通过登录接口获取 token
2. 每次请求都需要在 Authorization 头中携带 Bearer token
3. title 字段为必填项，不能为空
4. type 字段必须是 1（图文）或 2（视频）
5. 新创建的笔记默认状态为 0（审核中）
6. topic_ids、location 和 media_data 可以为空数组或 null
7. 图片上传建议先调用 `/v1/note/upload` 接口上传图片，获取 URL 后再创建笔记

---

# 点赞接口文档

## 接口列表

- POST /v1/note/:note_id/like （需要认证）
  - 说明：对指定笔记点赞
  - 成功响应：`{"code":200,"msg":"success","data":{"liked":true}}`

- DELETE /v1/note/:note_id/like （需要认证）
  - 说明：取消对指定笔记的点赞
  - 成功响应：`{"code":200,"msg":"success","data":{"liked":false}}`

- GET /v1/note/:note_id/like （需要认证）
  - 说明：查询当前用户是否已点赞该笔记
  - 成功响应：`{"code":200,"msg":"success","data":{"liked":true}}`

- GET /v1/note/:note_id/likes/count （无需认证）
  - 说明：查询指定笔记的点赞总数
  - 成功响应：`{"code":200,"msg":"success","data":{"like_count":123}}`

# 收藏接口文档

## 接口列表

- POST /v1/note/:note_id/collect （需要认证）
  - 说明：收藏指定笔记
  - 成功响应：`{"code":200,"msg":"success","data":{"collected":true}}`

- DELETE /v1/note/:note_id/collect （需要认证）
  - 说明：取消收藏指定笔记
  - 成功响应：`{"code":200,"msg":"success","data":{"collected":false}}`

- GET /v1/note/:note_id/collect （需要认证）
  - 说明：查询当前用户是否已收藏该笔记
  - 成功响应：`{"code":200,"msg":"success","data":{"collected":true}}`

- GET /v1/note/:note_id/collections/count （无需认证）
  - 说明：查询指定笔记的收藏总数
  - 成功响应：`{"code":200,"msg":"success","data":{"collect_count":123}}`

## 查询自己的收藏列表

- GET /v1/note/my/collects （需要认证）
  - 说明：分页查询当前用户收藏的笔记
  - 查询参数：`page`（默认1），`pagesize`（默认20，最大100）
  - 成功响应：`{"code":200,"msg":"success","data":{"notes":[...],"total":10}}`

## 请求头

```
Authorization: Bearer <token>   // 仅认证接口需要
Content-Type: application/json
```

## 说明

- 收藏状态记录在 `note_collections` 表（唯一键：note_id + user_id，status=1/0）。
- 收藏计数通过 `note_stats.coll_count` 维护，幂等且防负数。
- 接口响应结构与点赞保持一致：`{code, msg, data}`。

## 请求头

```
Authorization: Bearer <token>   // 仅认证接口需要
Content-Type: application/json
```

## 说明

- 点赞状态通过 `note_likes` 表记录（唯一键：note_id + user_id）。
- 点赞计数通过 `note_stats.like_count` 维护，支持幂等更新与防负数。
- 所有接口均返回统一响应结构：`{code, msg, data}`。

---

# 关注接口文档

## 查询关注列表

- GET `/v1/user/:user_id/following/list` （需要认证）
  - 说明：分页查询指定用户已关注的用户列表
  - 查询参数：`page`（默认1），`page_size`（默认20，最大100）
  - 成功响应示例：
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "list": [
        {
          "user_id": 123,
          "nickname": "张三",
          "avatar": "https://example.com/avatar.jpg",
          "updated_at": "2026-01-07 12:34:56"
        }
      ],
      "total": 1
    }
  }
  ```

---

# 笔记接口文档

## 接口列表

### 笔记上传与创建

#### POST /v1/note/upload
- **说明**: 上传笔记图片
- **认证**: 否
- **请求头**: `Content-Type: multipart/form-data`
- **请求参数**: 
  - `file` (file) - 图片文件
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "url": "https://example.com/images/xxxx.jpg",
      "thumbnail_url": "https://example.com/images/xxxx_thumb.jpg",
      "width": 1920,
      "height": 1080
    }
  }
  ```

#### POST /v1/note/create
- **说明**: 创建新的笔记（需要用户认证）
- **认证**: 是
- **请求头**: 
  ```
  Authorization: Bearer <token>
  Content-Type: application/json
  ```
- **请求体**:
  ```json
  {
    "title": "我的第一篇笔记",
    "content": "这是笔记的内容...",
    "topic_ids": [1001, 1002],
    "location": {
      "lat": 39.9042,
      "lng": 116.4074,
      "name": "北京天安门"
    },
    "media_data": [
      {
        "url": "https://example.com/image1.jpg",
        "thumbnail_url": "https://example.com/thumb1.jpg",
        "width": 1920,
        "height": 1080,
        "duration": 0
      }
    ],
    "type": 1,
    "visible_conf": 1
  }
  ```
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "note_id": 1234567890123456789
    }
  }
  ```

### 笔记查询

#### GET /v1/note/my
- **说明**: 获取当前用户的笔记列表
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **查询参数**:
  - `page` (int, 可选) - 页码，默认 1
  - `page_size` (int, 可选) - 每页数量，默认 20，最大 100
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "notes": [
        {
          "note_id": 123456,
          "title": "我的笔记",
          "content": "笔记内容...",
          "created_at": "2026-01-07 10:00:00",
          "like_count": 10,
          "collect_count": 5,
          "comment_count": 2
        }
      ],
      "total": 50
    }
  }
  ```

#### GET /v1/note/my/collects
- **说明**: 获取当前用户收藏的笔记列表
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **查询参数**:
  - `page` (int, 可选) - 页码，默认 1
  - `page_size` (int, 可选) - 每页数量，默认 20，最大 100
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "notes": [
        {
          "note_id": 123456,
          "user_id": 789,
          "nickname": "作者昵称",
          "avatar": "https://example.com/avatar.jpg",
          "title": "收藏的笔记",
          "content": "笔记内容...",
          "created_at": "2026-01-07 10:00:00",
          "like_count": 10,
          "collect_count": 5
        }
      ],
      "total": 20
    }
  }
  ```

### 点赞接口

#### POST /v1/note/:note_id/like
- **说明**: 对笔记点赞
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "liked": true
    }
  }
  ```

#### DELETE /v1/note/:note_id/like
- **说明**: 取消点赞
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "liked": false
    }
  }
  ```

#### GET /v1/note/:note_id/like
- **说明**: 查询当前用户是否已点赞该笔记
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "is_liked": true
    }
  }
  ```

#### GET /v1/note/:note_id/likes/count
- **说明**: 查询笔记的点赞总数
- **认证**: 否
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "like_count": 100
    }
  }
  ```

### 收藏接口

#### POST /v1/note/:note_id/collect
- **说明**: 收藏笔记
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "collected": true
    }
  }
  ```

#### DELETE /v1/note/:note_id/collect
- **说明**: 取消收藏
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "collected": false
    }
  }
  ```

#### GET /v1/note/:note_id/collect
- **说明**: 查询当前用户是否已收藏该笔记
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "is_collected": true
    }
  }
  ```

#### GET /v1/note/:note_id/collections/count
- **说明**: 查询笔记的收藏总数
- **认证**: 否
- **路径参数**: `note_id` - 笔记ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "collect_count": 50
    }
  }
  ```

---

# 关注接口文档

## 接口列表

### 关注操作

#### POST /v1/user/:user_id/follow
- **说明**: 关注指定用户
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `user_id` - 目标用户ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "followed": true
    }
  }
  ```
- **错误响应**:
  ```json
  {
    "code": 400,
    "message": "不能关注自己"
  }
  ```

#### DELETE /v1/user/:user_id/follow
- **说明**: 取消关注用户
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `user_id` - 目标用户ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "followed": false
    }
  }
  ```

#### GET /v1/user/:user_id/follow
- **说明**: 查询当前用户是否已关注该用户
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `user_id` - 目标用户ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "is_following": true
    }
  }
  ```

### 粉丝和关注统计

#### GET /v1/user/:user_id/followers/count
- **说明**: 查询用户的粉丝数
- **认证**: 否
- **路径参数**: `user_id` - 目标用户ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "follower_count": 100
    }
  }
  ```

#### GET /v1/user/:user_id/following/count
- **说明**: 查询用户的关注数
- **认证**: 否
- **路径参数**: `user_id` - 目标用户ID
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "following_count": 50
    }
  }
  ```

### 关注列表查询

#### GET /v1/user/:user_id/following/list
- **说明**: 查询用户已关注的用户列表
- **认证**: 是
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**: `user_id` - 目标用户ID
- **查询参数**:
  - `page` (int, 可选) - 页码，默认 1
  - `page_size` (int, 可选) - 每页数量，默认 20，最大 100
- **成功响应**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "list": [
        {
          "user_id": 123,
          "nickname": "张三",
          "avatar": "https://example.com/avatar.jpg",
          "updated_at": "2026-01-07 12:34:56"
        },
        {
          "user_id": 456,
          "nickname": "李四",
          "avatar": "https://example.com/avatar2.jpg",
          "updated_at": "2026-01-06 15:00:00"
        }
      ],
      "total": 50
    }
  }
  ```

## 错误状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（未登录或 token 无效） |
| 404 | 用户不存在 |
| 500 | 服务器内部错误 |

## 通用说明

- 所有需要认证的接口都需要在请求头中携带 `Authorization: Bearer <token>`
- token 从登录接口获取，有效期为 24 小时
- 分页接口默认返回 20 条数据，可通过 `page_size` 参数调整
- 所有时间字段使用格式: `YYYY-MM-DD HH:mm:ss`
