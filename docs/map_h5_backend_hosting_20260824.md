# H5 地图页后端托管

更新时间：2026-08-24

后端公开提供：

```text
https://www.hypercn.cn/map/index.html
```

Gin 路由为 `GET /map/*`，不经过 `/api` 鉴权。

## 生产部署

默认静态目录：

```text
/root/map-h5
```

必须包含：

```text
/root/map-h5/index.html
```

如需改目录，在 API 服务环境中设置：

```bash
MAP_H5_DIR=/your/path/map-h5
```

部署新版 API 二进制后，用浏览器访问 `/map/index.html` 验证。该页面调用同域的 `/api/v1/map/markers`，无需额外配置 CORS。

小程序仍需在微信公众平台配置 `https://www.hypercn.cn` 为业务域名，并将以下值写入小程序生产环境后重新构建：

```env
YDY_MAP_H5_URL=https://www.hypercn.cn/map/index.html
```

清空该变量并重新构建即可回退原生地图。
