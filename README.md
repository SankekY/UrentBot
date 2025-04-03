# **Scaut Bot — Telegram-бот для отслеживания активности скаутов**  

**Простой и легковесный Telegram-бот для учёта активности скаутов без использования базы данных.**  

---

## **📌 Особенности**  
- **Без базы данных** — Использует оперативную память (map-структуры)  
- **Чистый код** — Простота понимания и модификации  
- **Простые команды** — Базовый функционал для учёта работы  

---

## **🚀 Команды**  

| Команда | Описание |  
|---------|------------|  
| `/start` | Показать доступные команды |  
| `/info` | Показать текущую статистику сессии (без сброса) |  
| `/report` | Сформировать и отправить отчёт (сбрасывает сессию) |  
| `/rgl` | (Админ) Список активных скаутов |  
| `/stats` | (Админ) Глобальная статистика |  
| `/restats` | (Админ) Сбросить глобальную статистику |  
| `/sub` | Напоминание в ЛС, если сессия не закрыта за 1.5 часа |  

---

## **⚙️ Как это работает**  
- **Отслеживает**:  
  - **Перемещения** (`/report 12`)  
  - **Уборки** (загрузка фото)  
  - **Опоздания** (>30 минут между отчётами)  
- **Автосброс** через 90 минут неактивности  

---

## **🛠 Настройка**  
1. **Создайте бота** через [@BotFather](https://t.me/BotFather)  
2. **Установите токен** в `initConfig()`  
3. **Запустите**:  
   ```sh
   go run -race main.go
   ```

---

## **📊 Пример вывода**  
### **Режим скаута**  
```
👤 @ScoutName  
➖ Перемещений: 12  
➖ Уборок: 3  
➖ Опозданий: 0  
⏳ Сессия: 14:00 - 15:30  
```  

### **Режим RGL (Админ)**  
```
🟢 АКТИВНЫЕ СКАУТЫ  
@Scout1 → 12 перемещений, 3 уборки  
@Scout2 → 8 перемещений, 5 уборок  
```  

---

## **📝 Примечания**  
- **Нет сохранения данных** → Статистика сбрасывается при перезапуске бота  
- **Защита от гонок данных** → Используется `sync.RWMutex`  
- **Простой и быстрый** → Нет внешних зависимостей  

--- 

**🔹 Минимализм. Эффективность. Работоспособность.** 🚀