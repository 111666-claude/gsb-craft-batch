// Package craft 是制作系统：按配方消耗材料并产出成品。
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

// Craft 制作 count 份。批量是原子的：先整批校验材料，不够就一份不做、
// 库存分文不动；未解锁的配方返回 ErrLocked，同样不动库存。
// count <= 0 按 1 份处理。
func Craft(inventory map[string]int, recipe Recipe, count int) (int, error) {
	if count <= 0 {
		count = 1
	}
	if !recipe.Unlocked {
		return 0, ErrLocked
	}
	for material, amount := range recipe.Need {
		if inventory[material] < amount*count {
			return 0, ErrMaterial
		}
	}
	for material, amount := range recipe.Need {
		inventory[material] -= amount * count
	}
	return recipe.Output * count, nil
}
