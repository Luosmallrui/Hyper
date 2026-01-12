namespace go im.push

struct PushRequest {
    1: i64 cid
    2: i32 uid
    3: string payload
    4: string event
}

// 新增：批量推送请求
struct BatchPushRequest {
    1: list<i64> cids     // 同一个 Server 下的所有连接 ID
    2: i32 uid            // 目标用户 ID
    3: string payload     // 消息内容
    4: string event       // 事件类型（如 "chat"）
}

struct PushResponse {
    1: bool success
    2: string msg
}

service PushService {
    PushResponse PushToClient(1: PushRequest req)

    // 新增：批量推送接口
    PushResponse BatchPushToClient(1: BatchPushRequest req)
}