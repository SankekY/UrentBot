# **Scaut Bot - Telegram Bot for Scout Activity Tracking**  

**Simple, lightweight Telegram bot for tracking scout activity without a database.**  

---

## **📌 Features**  
- **No Database** - Uses in-memory storage (maps)  
- **Clean Code** - Easy to understand and modify  
- **Simple Commands** - Basic functionality for tracking scout work  

---

## **🚀 Commands**  

| Command | Description |  
|---------|------------|  
| `/start` | Show available commands |  
| `/info` | Display current session stats (without reset) |  
| `/report` | Generate and submit a work report (resets session) |  
| `/rgl` | (Admin) List all active scouts |  
| `/stats` | (Admin) Show global statistics |  
| `/restats` | (Admin) Reset global stats |  
| `/sub` | Get a reminder in DM if session isn't closed in 1.5h |  

---

## **⚙️ How It Works**  
- **Tracks**:  
  - **Movements** (`/report 12`)  
  - **Cleaning actions** (photo uploads)  
  - **Late reports** (>30min delay)  
- **Auto-reset** after 90min of inactivity  

---

## **🛠 Setup**  
1. **Create a bot** via [@BotFather](https://t.me/BotFather)  
2. **Set token** in `initConfig()`  
3. **Run**:  
   ```sh
   go run -race main.go
   ```

---

## **📊 Example Output**  
### **Scaut Mode**  
```
👤 @ScoutName  
➖ Moves: 12  
➖ Cleanups: 3  
➖ Late Reports: 0  
⏳ Session: 14:00 - 15:30  
```  

### **RGL Mode (Admin)**  
```
🟢 ACTIVE SCOUTS  
@Scout1 → 12 moves, 3 cleanups  
@Scout2 → 8 moves, 5 cleanups  
```  

---

## **📝 Notes**  
- **No persistence** → Stats reset on bot restart  
- **Race-safe** → Uses `sync.RWMutex`  
- **Simple & Fast** → No external dependencies  

--- 

**🔹 Minimalist. Efficient. Works.** 🚀