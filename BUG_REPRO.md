## Bug 是什么

预算创建、更新、存储隔离和请求校验的约定不一致。不存在的分类也能创建预算，年度预算请求被错误拒绝，更新周期后读取出的预算对象还能被调用方改坏。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run 'TestBudgetRequiresExistingCategoryAndKeepsPeriodUpdatesIsolated|TestBudgetRequestAcceptsYearlyPeriod' -count=1
```

## 错误信息

测试会稳定失败，典型信息包括：

```text
budget for missing category should fail
yearly budget should validate: period: must be monthly or yearly
```
