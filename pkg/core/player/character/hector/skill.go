package hector

import (
	"://github.com"
	"://github.com/attributes"
	"://github.com/combat"
)

func (c *char) Skill(p map[string]int) (int, int) {
	f, a := c.ActionFrames(core.ActionSkill, p)

	// 1. Первый удар Е-шки (211.4% Электро урон)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Skill Cast",
		AttackTag:  combat.AttackTagElementalSkill,
		ICDTag:     combat.ICDTagNone, // 0 ICD для гарантированной реакции
		Element:    attributes.Electro,
		Mult:       2.114,
	}

	// Очередь атаки с колбэком (проверяем, вызвана ли реакция)
	c.core.QueueAttack(ai, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 2), f, f, func(ac combat.AttackCB) {
		// Проверяем, закрыл ли удар Сверхпроводник или Звёздный проводник
		if ac.Reaction == core.Superconduct || ac.Reaction == core.StellarConduct {
			c.triggerSparkingTrace()
		}
	})

	c.QueueParticle("hector", 3, attributes.Electro, f+10)
	c.SetCD(core.ActionSkill, 12*60) // Откат 12 секунд (720 фреймов)

	return f, a
}

// Логика Искрящегося следа и Метки Холодной руки
func (c *char) triggerSparkingTrace() {
	// Доп. урон по площади (145.2%)
	aiArea := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Skill Area",
		AttackTag:  combat.AttackTagElementalSkill,
		Element:    attributes.Electro,
		Mult:       1.452,
	}
	c.core.QueueAttack(aiArea, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 4), 0, 0)

	// Запуск карманных тиков раз в 2.5 сек (150 фреймов) на протяжении 13 секунд
	for i := 1; i <= 5; i++ {
		delay := i * 150
		aiDot := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Sparking Trace",
			AttackTag:  combat.AttackTagElementalSkill,
			ICDTag:     combat.ICDTagElementalSkill,
			Element:    attributes.Electro,
			Mult:       0.687,
		}
		c.core.QueueAttack(aiDot, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 3), delay, delay, func(ac combat.AttackCB) {
			// С2: При попадании карманного тика баффаем всю пачку на 15% физ и стихий
			for _, char := range c.core.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithExpiry("hector-c2", 600), // 10 сек
					AffectedStat: attributes.PhyBonus,
					Amount: func() ([]float64, bool) {
						val := make([]float64, attributes.EndStatType)
						val[attributes.PhyBonus] = 0.15
						val[attributes.Electro] = 0.15
						val[attributes.Cryo] = 0.15
						return val, true
					},
				})
			}
		})
	}

	// Взрыв Метки Холодной руки через 3 секунды (180 фреймов) — Физ. урон (388.8%)
	aiMark := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Cold Hand Burst",
		AttackTag:  combat.AttackTagElementalSkill,
		Element:    attributes.Physical,
		Mult:       3.888,
	}
	c.core.QueueAttack(aiMark, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 5), 180, 180)
}
