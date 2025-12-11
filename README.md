# File Change Tracker

A lightweight server-side file integrity and change‑monitoring tool written in Go.  
This project currently includes a recursive `fsnotify` watcher and will expand to include hashing, baseline comparison, authorized‑user checks, and alerting.

---

## 🚀 Project Goals (High Level)

1. Maintain a **baseline** of important system files (SHA‑256 + metadata).
2. Monitor critical directories **recursively** and detect file creation, modification, deletion, and renaming in real time.
3. Distinguish **authorized** vs **unauthorized** modifications.
4. Log events in a structured form and raise alerts when unauthorized changes occur.
5. Run as a **systemd service** on Linux servers with automatic restart and reliability.
6. Provide a simple alerting mechanism (email/webhook/syslog).

---

## 📌 Current Status (What Exists Right Now)

The repository currently contains:

- `main.go` → Recursive `fsnotify` watcher that:
  - Walks all configured paths.
  - Adds all subdirectories to inotify.
  - Detects new directories and begins watching them automatically.
  - Logs file events like CREATE, WRITE, REMOVE.

- `go.mod`, `go.sum` → Go module metadata.

Hashing, baseline management, and authorization logic will be added next as separate modules.

---

## 🛠️ How to Set Up & Run (Local Development)

### 1️⃣ Clone the repository
```bash
git clone https://github.com/<your-username>/file-change-tracker.git
cd file-change-tracker


---

### 2️⃣ (Optional) Run watcher with custom paths
You can specify paths manually when running:

sudo go run main.go /etc /bin /sbin /usr/bin /usr/sbin /boot
---
###3️⃣ Build binary
go build -o filetracker main.go

---
4️⃣ Run binary
sudo ./filetracker /etc /bin /sbin /usr/bin /usr/sbin /boot 

You should now see output like:

watching: /etc
watching: /usr/bin
[CREATED] /etc/newfile.conf
[MODIFIED] /home/viz/test.txt
---
5️⃣ If you see inotify watch errors
Increase watch limit permanently:

echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

---
6️⃣ Testing changes
Open another terminal and try:

sudo mkdir /var/www/test-viz
sudo touch /var/www/test-viz/hello.txt
sudo rm -r /var/www/test-viz
---

-------------------------------------------------------------------------------------------------------------------------------------------------------------------------
## 🧩 Next Steps for Project 

These are the components we will build next:

### 🔹 Baseline Creator
- Scan all watched directories  
- Compute SHA‑256 hash for each file  
- Save results   
- Store: file path, permissions, owner, mod‑time, hash  

---

### 🔹 Event Handler Integration
Triggered on `WRITE` / `CREATE`:
1. Compute new hash  
2. Compare with baseline  
3. Decide whether to **update baseline** or **raise alert**  

---

### 🔹 Authorized-User Check
- **Initial method:** check file owner (`os.Stat`)  
- **Advanced method:** integrate `auditd` to map UID → username and detect which user performed the modification  

---

### 🔹 Baseline Update Policy
| Situation | Action |
|----------|--------|
| File changed by authorized user | Update baseline |
| File changed by unauthorized user | Alert + DO NOT update baseline |

This prevents attackers from modifying both file *and* baseline silently.

---

### 🔹 Logging & Alerts
- Log all events in `/var/log/filetracker/`  
- Use structured JSON logs  
- Add alert mechanisms:
  - Email  
  - Webhook (Discord / Slack)  
  - Syslog  
  - Telegram / SMS  

---

### 🔹 Systemd Service Packaging (for later)
- Auto‑start on boot  
- Restart on crash

---


