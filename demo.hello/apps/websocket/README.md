# WebSocket Demo

> frontend: `{js-project}/vue_pages/vue_apps`

描述：使用 websocket 实时同步后端任务状态。

场景：批量执行 jenkins job, 实时同步job状态，直到所有job执行完成。

并发访问：当有多个client与服务端建立websocket连接时，可以使用 eventbus 完成与多个client的数据同步。

