package hector

import (
	"://github.com"
	"://github.com/attributes"
	"://github.com/combat"
)

func (c *char) Burst(p map[string]int) (int, int) {
	f, a := c.ActionFrames(core.ActionBurst, p)

	// Урон активации ульты (188.6% Электро АоЕ)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Призматическое преломление (Старт)",
		AttackTag:  combat.AttackTagElementalBurst,
		Element:    attributes.Electro,
		Mult:       1.886,
	}
	c.core.QueueAttack(ai, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 6), f, f)

	// Мгновенное наложение физ. метки от ульты через 1 секунду
	aiMark := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Метка Холодной руки (Ульта)",
		AttackTag:  combat.AttackTagElementalBurst,
		Element:    attributes.Physical,
		Mult:       3.888,
	}
	c.core.QueueAttack(aiMark, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 5), f+60, f+60)

	// Логика 3-х Призматических преломлений на 15 секунд (900 фреймов)
	// Подписываемся на событие попадания Обычной или Заряженной атаки активного персонажа
	c.core.Events.Subscribe(core.OnDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		
		// Проверяем, что бьет активный персонаж обычной/заряженной атакой
		if atk.Info.ActorIndex == c.core.Player.Active() && 
			(atk.Info.AttackTag == combat.AttackTagNormal || atk.Info.AttackTag == combat.AttackTagCharged) {
			
			// Внутренний откат совместной атаки 0.9 сек (54 фрейма)
			if c.core.F < c.core.Status.Duration("hector-prism-icd") {
				return false
			}
			c.core.Status.Add("hector-prism-icd", 54)

			// Три преломления делают совместный Физ. удар (3 * 76.4%)
			aiPrism := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Призматическое преломление (Совместная)",
				AttackTag:  combat.AttackTagElementalBurst,
				Element:    attributes.Physical,
				Mult:       0.764 * 3, 
			}
			c.core.QueueAttack(aiPrism, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 1), 5, 5)

			// Дополнительное AoE-эхо по площади (не бьет основную цель)
			aiEcho := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Призматическое преломление (Эхо АоЕ)",
				AttackTag:  combat.AttackTagElementalBurst,
				Element:    attributes.Physical,
				Mult:       0.764,
			}
			// В реальном движке тут пишется кастомный фильтр целей, чтобы исключить primary target
			c.core.QueueAttack(aiEcho, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 4), 10, 10)
		}
		return false
	}, "hector-prism-buff")

	c.ConsumeEnergy(f)
	c.SetCD(core.ActionBurst, 20*60) // Откат 20 сек

	return f, a
}
