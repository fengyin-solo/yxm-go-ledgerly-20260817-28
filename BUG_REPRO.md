## Bug 是什么

交易删除、交易类型汇总和交易请求校验之间的契约不一致。删除借记交易后，账户余额没有恢复到删除前状态；已删除流水仍会出现在列表和月报计算中；零金额交易请求也没有被拒绝。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run 'TestTransactionDeleteRestoresBalanceAndRemovesLedgerEntry|TestMonthlyReportTreatsDebitsAsExpenses|TestTransactionRequestRejectsZeroAmount' -count=1
```

## 错误信息

测试会稳定失败，典型信息包括：

```text
deleted transaction still appears in list: 1
bad debit rollup: expense=42.25 net=-42.25 account=42.25
zero amount transaction should be rejected
```
