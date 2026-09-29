package craft

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
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

func TestBatchAtomicIncompleteIsAtomic(t *testing.T) {
	inventory := map[string]int{"ore": 5}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}, 3)
	if err != ErrMaterial {
		t.Fatalf("整批材料不够应该报 ErrMaterial：%v", err)
	}
	if crafted != 0 {
		t.Fatalf("不够就一份不做：%d", crafted)
	}
	if inventory["ore"] != 5 {
		t.Fatalf("失败不能动库存：%v", inventory)
	}
}

func TestBatchPartialMaterialShortage(t *testing.T) {
	inventory := map[string]int{"ore": 100, "wood": 2}
	recipe := Recipe{Name: "axe", Need: map[string]int{"ore": 1, "wood": 1}, Output: 1, Unlocked: true}
	crafted, err := Craft(inventory, recipe, 3)
	if err != ErrMaterial {
		t.Fatalf("部分材料不足应该报 ErrMaterial：%v", err)
	}
	if crafted != 0 {
		t.Fatalf("部分材料不足时一份都不做：%d", crafted)
	}
	if inventory["ore"] != 100 || inventory["wood"] != 2 {
		t.Fatalf("失败不能动任何材料：%v", inventory)
	}
}

func TestLockedRecipe(t *testing.T) {
	inventory := map[string]int{"ore": 9}
	crafted, err := Craft(inventory, Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: false}, 1)
	if err != ErrLocked {
		t.Fatalf("未解锁应该报 ErrLocked：%v", err)
	}
	if crafted != 0 {
		t.Fatalf("未解锁不能产出：%d", crafted)
	}
	if inventory["ore"] != 9 {
		t.Fatalf("未解锁不能动库存：%v", inventory)
	}
}

func TestRepeatSameRequest(t *testing.T) {
	recipe := Recipe{Name: "sword", Need: map[string]int{"ore": 3}, Output: 1, Unlocked: true}
	inventory := map[string]int{"ore": 5}
	first, err1 := Craft(inventory, recipe, 3)
	snapshot := inventory["ore"]
	second, err2 := Craft(inventory, recipe, 3)
	if first != second || err1 != err2 || inventory["ore"] != snapshot {
		t.Fatalf("同一批请求重复执行结果应该相同：(%d,%v) vs (%d,%v)", first, err1, second, err2)
	}
}

// referenceCraft 是参考实现：先整批校验再一次性扣减。
func referenceCraft(inventory map[string]int, recipe Recipe, count int) (int, error) {
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

func TestAgainstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 200; trial++ {
		kinds := 1 + rng.Intn(8)
		inventory := map[string]int{}
		need := map[string]int{}
		for i := 0; i < kinds; i++ {
			name := string(rune('a'+i)) + "ore"
			inventory[name] = rng.Intn(50)
			need[name] = 1 + rng.Intn(5)
		}
		recipe := Recipe{
			Name:     "rand",
			Need:     need,
			Output:   1 + rng.Intn(3),
			Unlocked: rng.Intn(2) == 0,
		}
		count := rng.Intn(12) - 2 // 覆盖 <=0 的语义

		gotInv := map[string]int{}
		refInv := map[string]int{}
		for k, v := range inventory {
			gotInv[k] = v
			refInv[k] = v
		}
		gotCrafted, gotErr := Craft(gotInv, recipe, count)
		refCrafted, refErr := referenceCraft(refInv, recipe, count)
		if gotCrafted != refCrafted || gotErr != refErr {
			t.Fatalf("第 %d 组产物或错误不一致：got (%d,%v) ref (%d,%v)", trial, gotCrafted, gotErr, refCrafted, refErr)
		}
		if len(gotInv) != len(refInv) {
			t.Fatalf("第 %d 组库存种类数不一致", trial)
		}
		for k, v := range refInv {
			if gotInv[k] != v {
				t.Fatalf("第 %d 组库存快照不一致：%s got=%d ref=%d", trial, k, gotInv[k], v)
			}
		}
	}
}

func TestScale(t *testing.T) {
	inventory := map[string]int{}
	need := map[string]int{}
	for i := 0; i < 100000; i++ {
		name := "mat" + strconv.Itoa(i)
		inventory[name] = 100000
		if i < 100 {
			need[name] = 1
		}
	}
	recipe := Recipe{Name: "bulk", Need: need, Output: 1, Unlocked: true}
	start := time.Now()
	crafted, err := Craft(inventory, recipe, 10000)
	elapsed := time.Since(start)
	if err != nil || crafted != 10000 {
		t.Fatalf("1 万份批量应该成功：%d %v", crafted, err)
	}
	if elapsed > time.Second {
		t.Fatalf("单次批量 1 万份应在 1 秒内完成：%v", elapsed)
	}
	for name := range need {
		if inventory[name] != 90000 {
			t.Fatalf("扣料总量应等于份数乘需求：%s=%d", name, inventory[name])
		}
	}
}
