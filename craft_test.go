package craft

import (
	"fmt"
	"math/rand"
	"testing"
)

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

func TestBatchAtomic(t *testing.T) {
	inventory := map[string]int{"ore": 5}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 3)
	if err != ErrMaterial || crafted != 0 {
		t.Fatalf("整批不够应该一份不做：crafted=%d err=%v", crafted, err)
	}
	if inventory["ore"] != 5 {
		t.Fatalf("失败不该扣料：%v", inventory)
	}
}

func TestLockedRecipe(t *testing.T) {
	inventory := map[string]int{"ore": 9}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: false}, 3)
	if err != ErrLocked || crafted != 0 {
		t.Fatalf("未解锁应该报 ErrLocked：crafted=%d err=%v", crafted, err)
	}
	if inventory["ore"] != 9 {
		t.Fatalf("未解锁不该扣料：%v", inventory)
	}
}

func TestPartialMaterialShortage(t *testing.T) {
	inventory := map[string]int{"ore": 100, "wood": 1}
	recipe := Recipe{Name: "spear", Need: map[string]int{"ore": 2, "wood": 2}, Output: 1, Unlocked: true}
	crafted, err := Craft(inventory, recipe, 2)
	if err != ErrMaterial || crafted != 0 {
		t.Fatalf("部分材料不足应该整批失败：crafted=%d err=%v", crafted, err)
	}
	if inventory["ore"] != 100 || inventory["wood"] != 1 {
		t.Fatalf("失败不该动任何材料：%v", inventory)
	}
}

func TestDeductEqualsCountTimesNeed(t *testing.T) {
	inventory := map[string]int{"ore": 30, "wood": 10}
	recipe := Recipe{Name: "spear", Need: map[string]int{"ore": 3, "wood": 1}, Output: 2, Unlocked: true}
	crafted, err := Craft(inventory, recipe, 4)
	if err != nil || crafted != 8 {
		t.Fatalf("产出应为 count×Output：crafted=%d err=%v", crafted, err)
	}
	if inventory["ore"] != 30-3*4 || inventory["wood"] != 10-1*4 {
		t.Fatalf("扣料总量应等于份数×需求：%v", inventory)
	}
}

// referenceCraft 是口径参考实现：先整批校验，再一次性扣减。
func referenceCraft(inventory map[string]int, recipe Recipe, count int) (int, error) {
	if count <= 0 {
		count = 1
	}
	if !recipe.Unlocked {
		return 0, ErrLocked
	}
	for material, amount := range recipe.Need {
		if amount > 0 && inventory[material] < amount*count {
			return 0, ErrMaterial
		}
	}
	for material, amount := range recipe.Need {
		if amount > 0 {
			inventory[material] -= amount * count
		}
	}
	return count * recipe.Output, nil
}

func cloneInventory(inventory map[string]int) map[string]int {
	cloned := make(map[string]int, len(inventory))
	for material, amount := range inventory {
		cloned[material] = amount
	}
	return cloned
}

func TestRandomCompareReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for round := 0; round < 200; round++ {
		kinds := 1 + rng.Intn(8)
		inventory := map[string]int{}
		for index := 0; index < kinds; index++ {
			inventory[fmt.Sprintf("mat%d", index)] = rng.Intn(50)
		}
		need := map[string]int{}
		for index := 0; index < 1+rng.Intn(kinds); index++ {
			need[fmt.Sprintf("mat%d", rng.Intn(kinds))] = rng.Intn(6)
		}
		recipe := Recipe{
			Name:     "random",
			Need:     need,
			Output:   1 + rng.Intn(3),
			Unlocked: rng.Intn(2) == 0,
		}
		count := rng.Intn(22) - 2 // 覆盖 count<=0 按 1 处理

		gotInventory := cloneInventory(inventory)
		wantInventory := cloneInventory(inventory)
		gotCrafted, gotErr := Craft(gotInventory, recipe, count)
		wantCrafted, wantErr := referenceCraft(wantInventory, recipe, count)

		if gotCrafted != wantCrafted || (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("第 %d 组产物不一致：got=(%d,%v) want=(%d,%v)", round, gotCrafted, gotErr, wantCrafted, wantErr)
		}
		if gotErr != nil && fmt.Sprint(gotInventory) != fmt.Sprint(inventory) {
			t.Fatalf("第 %d 组失败路径改动了库存：%v -> %v", round, inventory, gotInventory)
		}
		if len(gotInventory) != len(wantInventory) {
			t.Fatalf("第 %d 组库存规模不一致：%v vs %v", round, gotInventory, wantInventory)
		}
		for material, want := range wantInventory {
			if gotInventory[material] != want {
				t.Fatalf("第 %d 组库存不一致：%s got=%d want=%d", round, material, gotInventory[material], want)
			}
		}

		// 同一批请求重复执行结果相同。
		againInventory := cloneInventory(inventory)
		againCrafted, againErr := Craft(againInventory, recipe, count)
		if againCrafted != gotCrafted || (againErr == nil) != (gotErr == nil) ||
			fmt.Sprint(againInventory) != fmt.Sprint(gotInventory) {
			t.Fatalf("第 %d 组重复执行结果不同", round)
		}
	}
}

func TestScale(t *testing.T) {
	// 10 万种材料、单次批量 1 万份：耗时应只取决于配方材料种类数。
	inventory := make(map[string]int, 100000)
	for index := 0; index < 100000; index++ {
		inventory[fmt.Sprintf("mat%d", index)] = 100000
	}
	recipe := Recipe{
		Name:     "bulk",
		Need:     map[string]int{"mat0": 3, "mat99999": 7},
		Output:   1,
		Unlocked: true,
	}
	crafted, err := Craft(inventory, recipe, 10000)
	if err != nil || crafted != 10000 {
		t.Fatalf("1 万份批量应该成功：crafted=%d err=%v", crafted, err)
	}
	if inventory["mat0"] != 100000-3*10000 || inventory["mat99999"] != 100000-7*10000 {
		t.Fatalf("扣料总量不对：mat0=%d mat99999=%d", inventory["mat0"], inventory["mat99999"])
	}
}
