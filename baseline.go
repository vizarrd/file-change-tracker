package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

//
// 🔐 SINGLE SOURCE OF TRUTH
//
const baselinePath = "/var/lib/filetracker/baseline.json"

//
// -------- Data structures --------
//
type FileRecord struct {
	Hash  string `json:"hash"`
	Owner string `json:"owner"`
}

var baseline = make(map[string]FileRecord)

//
// -------- Ignore rules --------
//
var ignoreSuffixes = []string{
	".swp",
	".tmp",
	".bak",
	"~",
	".lock",
}

//
// -------- Logging helpers --------
//
func auditLog(format string, v ...interface{}) {
	log.Printf("[AUDIT] "+format, v...)
}

//
// -------- Baseline helpers --------
//
func shouldIgnore(path string) bool {
	lower := strings.ToLower(path)
	for _, suffix := range ignoreSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

func isBaselineEmpty() bool {
	return len(baseline) == 0
}

func loadBaseline() {
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		auditLog("Baseline not found, starting fresh")
		return
	}
	_ = json.Unmarshal(data, &baseline)
}

func saveBaseline() {
	// Ensure directory exists
	_ = os.MkdirAll(filepath.Dir(baselinePath), 0700)

	data, _ := json.MarshalIndent(baseline, "", "  ")
	_ = os.WriteFile(baselinePath, data, 0600)
}

func populateBaseline(paths []string) {
	for _, root := range paths {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if shouldIgnore(path) {
				return nil
			}
			hash, err := hashFile(path)
			if err != nil {
				return nil
			}
			baseline[path] = FileRecord{
				Hash:  hash,
				Owner: getFileOwner(info),
			}
			return nil
		})
	}
	saveBaseline()
}

//
// -------- File utilities --------
//
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getFileOwner(info os.FileInfo) string {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "unknown"
	}
	userObj, err := user.LookupId(fmt.Sprint(stat.Uid))
	if err != nil {
		return fmt.Sprint(stat.Uid)
	}
	return userObj.Username
}

//
// -------- Core change handler --------
//
func handleFileChange(path string) {
	if shouldIgnore(path) {
		return
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	newHash, err := hashFile(path)
	if err != nil {
		return
	}

	old, exists := baseline[path]
	if !exists {
		auditLog("🚨 UNAUTHORIZED NEW FILE DETECTED 🚨 %s", path)
		return
	}

	if old.Hash == newHash {
		return
	}

	// 🔥 STRICT MODE: ANY CHANGE IS UNAUTHORIZED
	auditLog("🚨 UNAUTHORIZED FILE MODIFICATION 🚨 %s", path)
}

//
// -------- Approval handler --------
//
func approveFile(path string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	hash, err := hashFile(path)
	if err != nil {
		return
	}

	baseline[path] = FileRecord{
		Hash:  hash,
		Owner: getFileOwner(info),
	}

	saveBaseline()
	auditLog("✅ APPROVED CHANGE %s", path)
}
