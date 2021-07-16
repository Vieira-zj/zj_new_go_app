# WebSocket Demo

> frontend: `{js-project}/vue_pages/vue_apps`

描述：使用 websocket 实时同步后端任务执行状态。

场景：批量执行 jenkins job, 实时同步job状态到client, 直到所有job执行完成。

测试点：

- 在数据同步过程中，同一个client重复多次发送数据同步消息。
- 并发访问：当有多个client与服务端建立websocket连接时，同时与多个client进行数据同步（使用 eventbus 完成消息发布与订阅）。

## APIs

Http:

- <http://localhost:8080/>
- <http://localhost:8080/mock/jobs>
- <http://localhost:8080/mock/jobs/init>

Websocket:

- <ws://localhost:8080/ws/echo>
- <ws://localhost:8080/ws/jobs/sync>

