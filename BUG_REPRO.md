## Bug 是什么

分类层级过滤和返回对象隔离在模型、服务、存储和 HTTP 查询处理之间不一致。只查询子分类时结果反了，读取出的子分类对象被调用方修改后会污染后续读取。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run 'TestCategoryHierarchyFilteringAndIsolation' -count=1
```

## 错误信息

测试会稳定失败，典型信息包括：

```text
child filter returned wrong categories
```
