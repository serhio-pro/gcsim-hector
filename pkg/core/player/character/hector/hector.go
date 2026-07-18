package hector

import (
	"://github.com"
	"://github.com/attributes"
	"://github.com/combat"
	"://github.com/player/character"
	"://github.com/player/character/profile"
)

// Описание структуры персонажа и его кастомных переменных
type char struct {
	*character.CharWrapper
	c6Stacks        int
	a4LastTriggered int
	core            *core.Core
}

// Инициализация Гектора в симуляторе
func init() {
	core.RegisterCharFunc(core.Hector, NewChar)
}

func NewChar(c *core.Core, char *character.CharWrapper, p profile.CharacterProfile) (character.Character, error) {
	s := char{
		CharWrapper: char,
		core:        c,
	}

	s.InitActionFrames()
	s.InitPassiveTalents()

	return &s, nil
}

// Фреймдата анимаций (Атака, Е-шка, Ульта в кадрах)
func (c *char) InitActionFrames() {
	// Обычная атака (Серия N5)
	c.AnimationFrames[core.ActionAttack] = []int{18, 15, 22, 19, 32} 
	// Элементальный навык (Е)
	c.AnimationFrames[core.ActionSkill] = 32
	// Взрыв стихий (Q)
	c.AnimationFrames[core.ActionBurst] = 56
}

// Логика пассивного таланта А1 (Леденящий резонанс) + Усиление от С6
func (c *char) InitPassiveTalents() {
	// Добавляем постоянный слушатель статов, который динамически баффает отряд
	c.core.Events.Subscribe(core.OnTick, func(args ...interface{}) bool {
		// Считываем текущую атаку Гектора
		currentAtk := c.Base.Atk*(1+c.Stats[attributes.ATK_P]) + c.Stats[attributes.ATK]
		
		if currentAtk > 1000 {
			excessAtk := currentAtk - 1000
			steps := int(excessAtk / 100)
			
			// Коэффициент: 1% без С6, 2% при наличии С6
			multiplier := 0.01
			maxBonus := 0.33
			if c.Base.Cons >= 6 {
				multiplier = 0.02
				maxBonus = 0.66
			}

			physBonus := float64(steps) * multiplier
			if physBonus > maxBonus {
				physBonus = maxBonus
			}

			// Проходим по всей пачке и раздаем Физ. бонус Крио, Электро и Пиро героям
			for _, char := range c.core.Player.Chars() {
				ele := char.Base.Element
				if ele == attributes.Cryo || ele == attributes.Electro || ele == attributes.Pyro {
					char.AddStatMod(character.StatMod{
						Base:         modifier.NewBase("hector-a1-buff", 1),
						AffectedStat: attributes.PhyBonus,
						Amount: func() ([]float64, bool) {
							val := make([]float64, attributes.EndStatType)
							val[attributes.PhyBonus] = physBonus
							return val, true
						},
					})
				}
			}
		}
		return false
	}, "hector-a1-monitor")
}

// Реализация Обычной Атаки и разгона критов от С6
func (c *char) Attack(p map[string]int) (int, int) {
	f, a := c.ActionFrames(core.ActionAttack, p)
	
	// Логика С6: добавляем стаки за каждый удар
	if c.Base.Cons >= 6 {
		c.c6Stacks++
		if c.c6Stacks > 10 {
			c.c6Stacks = 10
		}
		
		// Временный бафф критов на текущую атаку
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("hector-c6-crit", f),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				val := make([]float64, attributes.EndStatType)
				val[attributes.CR] = float64(c.c6Stacks) * 0.02
				val[attributes.CD] = float64(c.c6Stacks) * 0.04
				return val, true
			},
		})
	}

	// Генерация самого удара стрелы
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Normal Attack",
		AttackTag:  combat.AttackTagNormal,
		ICDTag:     combat.ICDTagNormalAttack,
		ICDGroup:   combat.ICDGroupDefault,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       0.924, // Множитель среднего удара
	}
	
	c.core.QueueAttack(ai, combat.NewCircleHit(c.core.Combat.PrimaryTarget(), 1), f, f)

	return f, a
}
