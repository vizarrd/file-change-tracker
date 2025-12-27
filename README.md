# File Change Tracker (File Integrity Monitoring Tool)

## Overview
File Change Tracker is a lightweight File Integrity Monitoring (FIM) tool
written in Go. It detects unauthorized file modifications using SHA‑256
hashing and updates trusted baselines only after explicit administrator
approval.

## Features
- Real‑time file change monitoring
- Cryptographic hash verification (SHA‑256)
- Strict unauthorized change detection
- Explicit admin approval workflow
- Clean audit‑only logging
- Runs as a systemd service (server‑style)

## Requirements
- Linux (Ubuntu recommended)
- Go 1.20+
- systemd
- Root privileges

## Installation

```bash
git clone https://github.com/<your-username>/file-change-tracker.git
cd file-change-tracker
go build -o filetracker
sudo cp filetracker /usr/local/bin/filetracker


##Setup Directories
sudo mkdir -p /var/lib/filetracker
sudo mkdir -p /var/log/filetracker
sudo chmod 700 /var/lib/filetracker
sudo chmod 700 /var/log/filetracker

##Run as a Service

sudo cp systemd/filetracker.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable filetracker
sudo systemctl start filetracker


##Usage

##View Audit Logs
sudo tail -f /var/log/filetracker/filetracker.log

##Approve a Legitimate Change
sudo systemctl stop filetracker
sudo filetracker --approve /path/to/file
sudo systemctl start filetracker


## How It Works (Architecture)

1. The tool runs as a background systemd service.
2. On startup, it builds or loads a trusted baseline containing:
   - File path
   - SHA‑256 hash
   - File owner
3. The service continuously monitors configured directories.
4. When a file is modified:
   - A new hash is calculated.
   - It is compared against the baseline.
5. If the hash differs:
   - The change is logged as **UNAUTHORIZED**.
   - The baseline is NOT updated.
6. Only an administrator can approve a change using `--approve`,
   which updates the baseline hash.


## Demo / Testing Workflow

1. Start the service:
```bash
sudo systemctl start filetracker

2.Monitor audit logs:

sudo tail -f /var/log/filetracker/filetracker.log

3.Modify a monitored file:

echo "unauthorized change" >> demo.txt

4.Observe unauthorized alert in logs.

5.Approve the change:

sudo systemctl stop filetracker
sudo filetracker --approve demo.txt
sudo systemctl start filetracker

6.Verify that the baseline hash is updated.


🎯 **Why this matters:**  
Anyone cloning your repo knows **exactly how to test it**.

---

## ✅ 3️⃣ Add “Configuration” section
Shows extensibility.

### 📌 Add this:

```md
## Configuration

Monitored directories can be modified inside `main.go`:

```go
var watchList = []string{
    "/etc",
    "/var/www",
    "/home/user/Desktop",
}
Temporary and editor files are ignored using suffix rules defined in
baseline.go.
=======
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
>>>>>>> 2146effe81551dba1f473de6851ace750974ec9c


---

<<<<<<< HEAD
=======
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

