package craft

import "testing"

func TestCraftOne(t *testing.T) {
	inventory := map[string]int{"ore": 3}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 1)
	if err != nil || crafted != 1 {
		t.Fatalf("一份制作应该成功：%d %v", crafted, err)
	}
	if inventory["ore"] != 0 {
		t.Fatalf("材料应该扣光：%v", inventory)
	}
}

func TestNotEnoughMaterial(t *testing.T) {
	inventory := map[string]int{"ore": 1}
	if _, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 1); err != ErrMaterial {
		t.Fatalf("材料不足应该报 ErrMaterial：%v", err)
	}
	if inventory["ore"] != 1 {
		t.Fatalf("材料不足时不该扣料：%v", inventory)
	}
}

func TestZeroCountMeansOne(t *testing.T) {
	inventory := map[string]int{"ore": 3}
	crafted, _ := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 0)
	if crafted != 1 {
		t.Fatalf("count=0 按 1 处理：%d", crafted)
	}
}

func TestBatchCount(t *testing.T) {
	inventory := map[string]int{"ore": 9}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 3)
	if err != nil || crafted != 3 {
		t.Fatalf("三份应该都做出来：%d %v", crafted, err)
	}
}
