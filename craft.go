// Package craft 是制作系统：按配方消耗材料并产出成品。
// 缺陷：材料校验与扣减分开做（失败留下半成品）、失败不返还、未解锁配方也能做。
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

// Craft 制作 count 份。缺陷：逐份扣料，中途不够就直接返回（前面几次的料已经扣掉了）。
func Craft(inventory map[string]int, recipe Recipe, count int) (int, error) {
	if count <= 0 {
		count = 1
	}
	crafted := 0
	for index := 0; index < count; index++ {
		for material, amount := range recipe.Need {
			if inventory[material] < amount {
				return crafted, ErrMaterial // 缺陷：已经扣掉的料不返还
			}
			inventory[material] -= amount
		}
		crafted += recipe.Output
	}
	return crafted, nil
}
