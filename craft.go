// Package craft 是制作系统：按配方消耗材料并产出成品。
// 批量制作是原子的：先整批校验，全部通过才一次性扣减；任何失败路径都不改动库存。
package craft

import "errors"

// ErrMaterial 表示材料不足。
var ErrMaterial = errors.New("craft: 材料不足")

// ErrLocked 表示配方还没解锁。
var ErrLocked = errors.New("craft: 配方还没解锁")

// Recipe 是一张配方。
type Recipe struct {
	Name     string
	Need     map[string]int
	Output   int
	Unlocked bool
}

// Craft 制作 count 份。count <= 0 按 1 份处理。
// 未解锁返回 ErrLocked；任一种材料不够 count 份返回 ErrMaterial；
// 两种失败都不改动库存。成功时扣料总量 = count × 配方需求。
func Craft(inventory map[string]int, recipe Recipe, count int) (int, error) {
	if count <= 0 {
		count = 1
	}
	if !recipe.Unlocked {
		return 0, ErrLocked
	}
	// 先整批校验：只读库存，O(配方材料种类)，不扫描整张库存表。
	// 用除法比较避免 amount*count 溢出。
	for material, amount := range recipe.Need {
		if amount <= 0 {
			continue
		}
		if inventory[material]/amount < count {
			return 0, ErrMaterial
		}
	}
	// 全部够用才一次性扣减。
	for material, amount := range recipe.Need {
		if amount <= 0 {
			continue
		}
		inventory[material] -= amount * count
	}
	return count * recipe.Output, nil
}
