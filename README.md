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

Result:
Unauthorized changes are detected, logged, and only trusted after admin approval.

## Configuration

Monitored directories can be modified inside `main.go`:

var watchList = []string{
    "/etc",
    "/var/www",
    "/home/user/Desktop",
}


Temporary and editor files are ignored using suffix rules defined in
baseline.go.

##Security Model
Default‑deny: all file changes are considered unauthorized

Baseline updates occur only via admin approval

No automatic trust updates

Audit logs contain security‑relevant events only

##License
This project is released under the MIT License.
