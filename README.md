# craft-batch

制作系统：按配方批量消耗材料并产出成品，只用 Go 标准库。

```
go test ./...
go run ./cmd/craft-batch --sample atomic
go run ./cmd/craft-batch --sample unlock
```

## 口径（README 为准）

- **批量制作要原子**：先校验 count 份所需材料是否都够，不够就一份都不做、材料分文不动。
  `--sample atomic`（5 个矿、配方每份 3 个、要做 3 份）按口径是 `crafted=0 ore_left=5`，
  现在做了 1 份再失败，给 `crafted=1 ore_left=2`。
- **未解锁配方不能做**：`--sample unlock` 按口径是 `crafted=0 ore_left=9`（并返回 `ErrLocked`），
  现在照样做出 1 份。
- 不变量：任何失败路径都不能改变库存；成功时扣料总量 = 产出份数 × 配方需求；
  同一批请求重复执行结果相同。
- 规模：10 万种材料、单次批量 1 万份 1 秒内完成，内存 O(材料种类)；不许逐份扫描整张库存表。

## 现在的行为

- `Craft` 逐份扣料，中途不足就直接 return：已经扣掉的料不返还。
- `Recipe.Unlocked` 完全没被读过。

## 输出

```
crafted=0 err=true ore_left=5
crafted=0 err=true ore_left=9
```

## 目录

```
craft.go            制作与原子性
cmd/craft-batch     命令行入口
craft_test.go       go test 用例
```
