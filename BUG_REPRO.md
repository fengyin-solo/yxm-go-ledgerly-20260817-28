## Bug 是什么

交易筛选在服务层、模型过滤和 HTTP 参数传递之间不一致。按账户、类型和日期边界一起筛选时，月初边界会被排除，账户过滤错用分类字段，类型过滤也可能被绕过。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run 'TestTransactionFilterHonorsInclusiveDatesTypeAndAccount' -count=1
```

## 错误信息

测试会稳定失败，典型信息包括：

```text
filter returned wrong transactions
```
