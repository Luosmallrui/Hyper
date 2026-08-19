package adapter

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// wsWriteTimeout 单次写操作超时，防止慢客户端阻塞推送协程
	wsWriteTimeout = 10 * time.Second
	// wsReadLimit 单条消息最大体积，防止超大消息耗尽内存
	wsReadLimit = 1 << 20 // 1MB
)

// WsAdapter Websocket 适配器
type WsAdapter struct {
	conn *websocket.Conn
}

var defaultUpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWsAdapter(w http.ResponseWriter, r *http.Request) (*WsAdapter, error) {

	conn, err := defaultUpGrader.Upgrade(w, r, w.Header())
	if err != nil {
		return nil, err
	}

	conn.SetReadLimit(wsReadLimit)

	return &WsAdapter{conn: conn}, nil
}

func (w *WsAdapter) Network() string {
	return NetworkWss
}

func (w *WsAdapter) Read() ([]byte, error) {
	_, content, err := w.conn.ReadMessage()
	return content, err
}

func (w *WsAdapter) Write(bytes []byte) error {
	if err := w.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout)); err != nil {
		return err
	}
	return w.conn.WriteMessage(websocket.TextMessage, bytes)
}

func (w *WsAdapter) Close() error {
	return w.conn.Close()
}

func (w *WsAdapter) SetCloseHandler(fn func(code int, text string) error) {
	w.conn.SetCloseHandler(fn)
}
