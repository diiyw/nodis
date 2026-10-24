# Nodis

Nodis 是一个嵌入式优先、可采用客户端/服务端部署的文档型数据库。

- Document 是命名的数据容器。
- Object 是 Document 中由动态字段组成的记录。
- 字段类型在写入时推断，无需预先声明 Schema。
- Database、Document、Object、Field 均可配置访问策略。
- Object 可绑定事务型执行管线；任何阶段失败都会回滚。
- WAL、对象编码和同步事件都使用 Nodis 自有字节格式。
- 客户端可将有权限的数据同步成本地物化视图。

当前内核已实现持久化快照、WAL 回收、定时检查点、自动字段二级索引和 TLS 1.3 / 双向 TLS。每个 Object 最多有 65,535 个字段；写入或管线产出超过此上限时，提交在写入 WAL 前返回 `ErrInvalidValue`。

```go
db, _ := nodis.Open(nodis.Options{Dir: "./data"})
defer db.Close()

admin := nodis.Principal{ID: "admin", Roles: []string{"admin"}}
doc, _ := db.CreateDocument(admin, "users")

tx := db.Begin(admin)
draft, _ := tx.Create(doc.ID())
_ = draft.Set("name", "Ada")
_ = draft.Set("age", int64(37))
id, _ := draft.Save()
_ = tx.Commit()

object, _ := db.Get(admin, doc.ID(), id)
name, _ := object.String("name")
_ = name
```

详细设计见 [DESIGN.md](DESIGN.md)。

## 条件查询、投影与分页

```go
adult, err := nodis.Filter("age", nodis.GreaterEqual, int64(18))
query := nodis.Query{
    Document: doc.ID(),
    Where: []nodis.Condition{adult},
    Select: []string{"name", "age"},
    Limit: 50,
}
page, err := db.Find(ctx, admin, query)
for i := range page.Objects {
    name, _ := page.Objects[i].String("name")
    _ = name
}
if page.HasMore {
    query.After = page.Next
    // 下一次 Find 从这个 Object ID 之后继续。
}
```

`Where` 中所有条件必须同时成立。支持 `Equal`、`NotEqual`、`Less`、`LessEqual`、`Greater`、`GreaterEqual`、`Exists`、`Missing`、`Prefix`。`Prefix` 用于字符串和字节；存在性条件可直接写 `nodis.Condition{Field: "email", Op: nodis.Exists}`。`Select` 为空时返回所有可读字段；指定不存在的字段会被忽略。

比较要求值类型一致，不同类型的 `NotEqual` 也不匹配。各种有符号整数宽度均推断为 KindInt，但有符号整数、无符号整数与浮点数彼此区分。缺少字段不会满足普通值比较；Null 与缺失不同。查询条件不接受数组、嵌套对象或 NaN，布尔值与 Null 只支持相等性比较。

查询需要 Document 和每个结果 Object 的 Discover / Read 权限。条件涉及的字段必须可读，无权读取时该对象不匹配该条件，包括 `Exists`、`Missing` 和 `NotEqual`；投影也会继续裁剪隐藏字段。隐藏对象不会影响 `HasMore`，查询不返回全库总数。成功查询会产生 DocumentRead 和各结果的 ObjectRead 事件。

分页按 Object ID 升序，`After` 为排他的 ID 游标，即使该对象被删除仍可继续。只有 `HasMore` 为真时 `Next` 才为最后一个结果的 ID，否则为 0。默认每页 100 条，最多 1024 条；最多 64 个条件、128 个投影字段。单页还限制为 8 MiB 的投影对象字节，可能提前结束一页；单个投影对象超过限制时返回 `ErrQueryTooLarge`，可缩小投影后再查询。

每页提供 `Sequence` 和 `PermissionRevision`，代表该页读取时的已提交状态。跨页不保留 MVCC 快照，中途修改与权限变化会在下一页体现；晚提交且 ID 小于已返回游标的对象不会出现在后续页中。

需要复用结果内存时：

```go
var buffer nodis.QueryBuffer
page, err := db.FindInto(ctx, admin, query, &buffer)
```

预热后的 `FindInto` 实测 **0 allocs**，包含条件过滤、投影和读取事件发布。返回的对象字节借用缓冲区，有效期到该缓冲区下一次使用；并发查询须各用一个缓冲区。普通 `Find` 返回可独立持有的结果。

每个 Document 自动维护泛型有序主索引及字段二级索引，无需创建索引或声明字段类型。同一字段的不同值类型在索引中分别排序；支持等值、不等值、范围和字符串／字节前缀定位。`Exists`、`Missing` 继续扫描；数组、嵌套对象和 NaN 不进入标量索引，但仍满足存在性判断。

多个 AND 条件会选择候选数最少的单字段索引；若游标之后的主索引更短，则使用主索引。候选 Object ID 重新按升序排列，并重新检查每个条件及对象／字段权限，之后才生成投影、读取事件和分页标记。索引候选数量不向客户端公开。查询期间仍持有数据库读锁，客户端 `Find` / `FindAsync` 会自动使用服务端索引。

事务成功写入 WAL 后，在同一提交锁内更新对象和索引；回滚、管线拒绝、权限拒绝和 WAL 写入失败不改变索引。恢复先还原快照并重放 WAL，再批量构建最终存活对象的索引，因此兼容已有磁盘格式，也不会索引已删除的历史对象。

二级索引使用泛型 AVL 树，节点集中存储在可复用的 arena 中，按值和 Object ID 排序。键借用不可变对象字节，避免重复复制字符串／字节。子树计数提供 O(log n) 边界定位及准确的候选数量，有序链提供 O(k) 的范围遍历。已预留容量或复用空闲节点时，插入／删除为 O(log n)，并实测 **0 allocs**；扩容仍可能分配内存和复制 arena。恢复时排序最终存活条目，再直接构建平衡树，避免逐条旋转插入。

树节点比原连续数组条目占用更多内存；当前 amd64 构建的标量节点为 64 字节，原条目为 40 字节，均不含容器和预留容量。空闲节点清除借用键引用，再供后续写入复用。主索引仍为有序数组，删除 Object ID 仍可能移动 O(n) 个 ID。候选 ID 缓冲区最多随所选范围增长，排序为 O(k log k)。尚未实现复合索引、多索引交集、索引开关、自定义排序和跨页 MVCC 快照。

本机索引微基准（Intel i5-11500，linux/amd64，Go 1.26；每次随机移除旧值再插入新值，固定条目数量，容量已准备）：

| 条目数 | 原有序数组 | AVL 树 | 两者分配 |
| --- | --- | --- | --- |
| 1,024 | 0.949 微秒/次 | 0.481 微秒/次 | 0 B/op，0 allocs/op |
| 16,384 | 21.303 微秒/次 | 0.738 微秒/次 | 0 B/op，0 allocs/op |
| 131,072 | 775.018 微秒/次 | 1.568 微秒/次 | 0 B/op，0 allocs/op |

同一轮 10,000 个对象的预热内核等值查询为 0.498 微秒/次，0 allocs/op。这些数据只测索引维护或内核查询，不包含完整事务、WAL、网络和 TLS；可复现命令为 `go test -run '^$' -bench '^BenchmarkSecondaryIndex(Mutation|FindInto)$' -benchmem -benchtime=150ms -count=1`。性能随机器和数据分布变化。

## 客户端/服务端

服务端使用 NDC1 原生字节协议。认证器只负责把凭据转换为 Principal，所有权限仍由数据库在服务端判断：

```go
listener, _ := net.Listen("tcp", "127.0.0.1:6380")
s := server.New(db, server.StaticTokens{
    "token": {ID: "ada", Roles: []string{"writer"}},
})
go s.Serve(listener)

remote, _ := client.Dial(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{
    Documents: []nodis.DocumentID{doc.ID()},
})
defer remote.Close()

id, err := remote.Create(ctx, doc.ID(), map[string]any{"name": "Ada"})
cached, ok := remote.Cached(doc.ID(), id)
```

`Create`、`Update` 和 `Delete` 成功时，服务端已经完成鉴权、对象管线和 WAL 提交。异步 API 返回泛型 `Future[T]`，完成语义相同。`Cached` 只读取经过字段权限裁剪的本地物化视图，通过独立连接最终更新；写入返回成功时缓存可能尚未追上。

仅需请求服务端、无需自动同步时：

```go
remote, err := client.DialWithOptions(ctx, address, credential, client.DialOptions{
    DisableSync: true,
})
```

此模式不创建订阅，`Cached` 保持为空；无需同步权限即可执行已获授权的普通操作。

进程内与远程客户端均支持服务端权威查询：

```go
page, err := remote.Find(ctx, query)
future := remote.FindAsync(ctx, query)
page, err = future.Wait(waitCtx)
```

查询通过 NDC1 的 `OpFind` 操作传输 NDQ1 条件和 NQP1 结果页，TLS 连接同样适用。`Find` 不依赖同步缓存；异步查询会复制条件及投影输入，返回后调用方可以复用原始切片。进程内客户端还提供 `FindInto`。

## 可靠同步与请求取消

进程内客户端和 TCP 客户端都提供 `SyncStatus()`：

```go
status := remote.SyncStatus()
// Enabled：是否开启自动同步。
// Ready：是否已完整接收并发布初始或重建快照。
// Sequence / PermissionRevision：已接收事件的序号水位与权限版本。
// Err：最近的同步错误；重建成功后清除。
```

检测到同步连接断开、事件损坏、订阅队列溢出或权限变更后，客户端立即清空可见缓存。重连快照先暂存，只有匹配序号和权限版本的结束标记到达后才整体发布。半途失败的快照不会暴露部分数据；离线期间删除的对象或失去权限的字段也不会残留。临时失去订阅权限后会退避重试，重新授权后可自行恢复。

`Ready` 表示快照完整，不能用作某次写入已同步到本地的确认。权限裁剪会跳过部分数据库事件，`Sequence` 也不是全库提交屏障。当前重连使用完整快照，尚无持久化续传游标或心跳探测；静默网络故障需要等待传输层报告错误。已返回给调用方的对象仍是可持有的历史值，权限撤销不会修改这些引用。

普通请求在排队等待连接时也响应 Context 取消，取消后不会发送请求。已发送请求遇到取消、传输错误或协议损坏时，客户端丢弃命令连接，下一次独立请求重新连接并认证。客户端不会自动重发失败的操作。

```go
id, err := remote.Create(ctx, document, fields)
if errors.Is(err, client.ErrOutcomeUnknown) {
    // 服务端可能已经提交。先核对业务状态，再决定是否提交新操作。
}
```

`OutcomeUnknownError` 同时保留底层错误，可用 `errors.Is(err, context.Canceled)` 等判断原因。取消客户端等待不等于取消服务端事务。异步操作的 Future 也保留这一语义；如果只取消 `Future.Wait` 的等待 Context，后台操作仍可能继续。关闭客户端会解除网络等待、取消重连并清空缓存，关闭后拒绝新的操作。

## 快照与恢复

```go
db, err := nodis.Open(nodis.Options{
    Dir: "./data",
    Durable: true,
    CheckpointInterval: 5 * time.Minute,
})
// 也可以手动执行；错误应由调用方处理。
err = db.Checkpoint()
// 获取最近一次手动或定时检查点的结果。
err = db.CheckpointError()
```

设置 `Dir` 后，默认在返回每次提交成功前同步 WAL，无需等待检查点。旧的 `Durable` 字段保留源码兼容，但 `Durable: false` 不再关闭同步；只有显式设置 `UnsafeNoSync: true`（服务端为 `-durable=false`）才跳过提交刷盘，此时断电或系统崩溃可能丢失最近的已确认写入。`Durable: true` 与 `UnsafeNoSync: true` 不能同时设置。`Flush()` 和 `Close()` 仍会同步并返回错误，调用方必须检查这些错误。`Dir` 为空时仍是纯内存数据库。

打开数据库时会同步 WAL 文件、数据目录及其祖先目录，确保新建文件和多层目录的目录项落盘。数据库在整个打开期间持有 `nodis.lock` 独占文件锁，同一目录的第二个实例会返回 `storage.ErrLocked`；进程退出后系统自动释放锁，运行期间不要删除锁文件。当前支持具有目录同步和 advisory lock 能力的本地文件系统（Linux、macOS、FreeBSD、OpenBSD、NetBSD、DragonFly BSD）；其他平台的持久化打开会报错，不会静默跳过锁保护。

WAL 写入或同步失败后，数据库拒绝后续写入，必须关闭并重新打开恢复。失败提交的完整记录可能已经落盘，因此报错不代表一定未提交，恢复后应核对状态再重试。恢复只丢弃不完整的最后一条记录，并同步修复结果；校验和错误等损坏会报错，不会自动截断掩盖。磁盘格式与现有数据兼容。可靠性依赖文件系统及设备正确兑现同步语义，磁盘损坏仍需备份或复制保护。

`CheckpointInterval: 0` 禁用定时检查点，嵌入式 API 默认禁用；手动快照仍可使用。纯内存数据库不支持检查点。

快照保留完整字段字典、ACL、权限版本、对象原始字节、管线绑定和 ID 计数器，包含已删除字段的字典项，避免权限引用发生偏移。应用需在重新打开后重新注册管线执行代码。

快照临时文件同步并原子发布、目录同步完成后，才截断 WAL。重启先恢复快照，再重放较新的 WAL；快照发布后、WAL 回收前发生中断也可恢复。记录校验和、终止标记及包含记录数与累计校验和的尾部共同检测不完整或损坏快照。

当前检查点持有数据库写锁，会阻塞写入及需要读锁的查询。活跃对象仍全部驻留内存，尚未实现磁盘按需读取、mmap 和后台分段压缩。

## TLS 部署

```go
cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
go s.ServeTLS(listener, &tls.Config{Certificates: []tls.Certificate{cert}})

remote, err := client.DialTLS(ctx, address, credential, scope, &tls.Config{
    RootCAs: trustedRoots,
    ServerName: "db.example.com",
})
```

请求连接和同步连接均使用至少 TLS 1.3，并验证服务端证书。双向 TLS 在服务端设置 `ClientAuth: tls.RequireAndVerifyClientCert` 与 `ClientCAs`，客户端设置 `Certificates`。双向 TLS 启用后仍需凭据认证，凭据由认证器映射为权限主体。

命令行示例（`NODIS_TOKEN` 为必需的凭据环境变量）：

```sh
go run ./cmd/nodis-server -listen 127.0.0.1:6380 -data ./data \
  -durable=true -checkpoint-interval=5m \
  -tls-cert=server.crt -tls-key=server.key
```

加上 `-tls-client-ca=clients-ca.crt` 可强制双向证书验证。服务端 CLI 默认每五分钟执行检查点，`-checkpoint-interval=0` 禁用。未配置 TLS 证书时使用普通 TCP。

## 泛型与分配约束

存储帧通过 `FrameWriter[W]` / `FrameReader[R]` 泛型实现；网络帧通过 `protocol.CommandWriter[W]` / `CommandReader[R]` 泛型实现。两者均复用缓冲区，读取结果借用底层缓冲区，有效期到下一次读取。NDC1 字节格式保持兼容，客户端将缓存对象复制为独立字节，避免帧缓冲区复用影响已返回结果。

回归测试验证已有对象的标量字段查找、缓存查询、同步状态读取、预分配协议编码、预热后的帧读写及文件 WAL 追加为 **0 allocs**，包括真实 TCP 连接上的帧往返。完整 CRUD、缓存对象更新、快照构建、TLS 建连和调用方持有的网络结果仍会分配内存。

查询的零分配检查覆盖：复用缓冲区的索引选择、等值／不等值／范围／前缀查询、投影和事件发布，预留容量的泛型主索引操作、AVL 增删／边界定位／范围游标、紧凑对象索引维护、查询/结果页编解码，以及服务端查询处理与响应编码。首次构建索引、arena／候选／结果缓冲区扩容、普通 `Find`、远程客户端持有的结果页、条件构建和异步 Future 仍允许分配。可用 `go test -run '^$' -bench '^BenchmarkSecondaryIndexFindInto$' -benchmem` 查看 1,000 / 10,000 个对象的查询基准。

```sh
go test ./...
go test -race ./...
go vet ./...
```
