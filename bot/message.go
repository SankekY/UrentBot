package bot

import (
	"fmt"
	"strings"
	"time"
)

func (b *Bot) GenerateReportScaut(scaut Scaut) string {
	var result strings.Builder
	if scaut.UserName != "" {
		scaut.TimeStart = scaut.TimeStart.Add(15 * time.Minute)
		sumHour := strings.Split(scaut.TimeStart.Sub(scaut.FirstTime).String(), "h")
		result.WriteString(fmt.Sprintf(
			"Смену завершил %s.c %s-%s (%s Часов)\n"+
				"🔁Перемещения: %d\n"+
				"✅Навёл порядок: %d\n"+
				"Итого: %d\n",
			getDate(), b.getTimeReport(scaut.FirstTime),
			b.getTimeReport(scaut.TimeStart), sumHour[0],
			scaut.Moved, scaut.Images, scaut.Moved+scaut.Images,
		))
	}
	return result.String()
}

func (b *Bot) GenerateReportRGL(scaut Scaut) string {
	var result strings.Builder
	if scaut.UserName != "" {
		kpd := efficiencyPercent(float64(scaut.TimeStart.Sub(scaut.FirstTime).Hours()), float64(scaut.Moved))
		result.WriteString(fmt.Sprintf(
			"👤 *@%s*\n"+
				"Смену завершил %s.c %s-%s (%s Часов)\n"+
				"🔁 Перемещения: %d\n"+
				"✅ Навёл порядок: %d\n"+
				"⏱ Отчёты более 30 минут: %d\n"+
				"🚀 Эффективность: %.1f%%\n\n",
			scaut.UserName,
			getDate(), b.getTimeReport(scaut.FirstTime),
			b.getTimeReport(scaut.TimeStart), scaut.TimeStart.Sub(scaut.FirstTime).String(),
			scaut.Moved, scaut.Images, scaut.Lateness, kpd,
		))
	}
	return result.String()
}

func (b *Bot) RGLStats(scouts map[int64]Scaut) string {
	var result strings.Builder
	result.WriteString("📊 *Текущая активность скаутов:*\n\n")

	for _, scout := range scouts {
		if !scout.FirstTime.IsZero() {
			kpd := efficiencyPercent(float64(scout.TimeStart.Sub(scout.FirstTime).Hours()), float64(scout.Moved))
			result.WriteString(fmt.Sprintf(
				"👤 *@%s*\n"+
					"➖ Перемещений: %d\n"+
					"➖ Уборок: %d\n"+
					"➖ Опозданий: %d\n"+
					"⏳ Время работы: %s - %s\n"+
					"🚀 Эффективность: %.1f%%\n\n",
				scout.UserName,
				scout.Moved,
				scout.Images,
				scout.Lateness,
				b.getTimeReport(scout.FirstTime),
				b.getTimeReport(scout.TimeStart),
				kpd,
			))
		}
	}

	result.WriteString("_Последнее обновление: " + time.Now().Format("15:04") + "_")
	return result.String()
}

func GenerateStats(Stats map[int64]WendayScaut) string {
	var result strings.Builder
	result.WriteString("📊 *Общая статистика скаутов!*\n\n")

	for _, scout := range Stats {
		if scout.UserName != "" {
			kpd := efficiencyPercent(float64(scout.SummerHour), float64(scout.SummerMuved))
			result.WriteString(fmt.Sprintf(
				"👤 *@%s*\n"+
					"➖ Перемещений: %d\n"+
					"➖ Опозданий: %d\n"+
					"⏳ Время работы: (%d Чаосв)\n"+
					"🚀 Эффективность: %.1f%%\n\n",
				scout.UserName,
				scout.SummerMuved,
				scout.SummerLateness,
				scout.SummerHour,
				kpd,
			))
		}
	}
	return result.String()
}
