namespace go im.push

struct PushRequest {
    1: i64 cid
    2: i32 uid
    3: string payload
    4: string event
}

struct PushResponse {
    1: bool success
    2: string msg
}

service PushService {
    PushResponse PushToClient(1: PushRequest req)
}