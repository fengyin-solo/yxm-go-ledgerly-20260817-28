## Bug 是什么

转账路径在余额边界、收款账户入账、收款方流水查询和请求金额校验上不一致。刚好转出账户全部余额时会被误判为余额不足，小额正数请求被校验拒绝，收款方也查不到完整历史。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run 'TestTransferExactBalanceCreditsDestinationAndHistory|TestTransferValidationKeepsSmallPositiveAmounts' -count=1
```

## 错误信息

测试会稳定失败，典型信息包括：

```text
exact balance transfer should succeed: invalid input: insufficient funds
small positive transfer should validate: amount: must be positive
```
