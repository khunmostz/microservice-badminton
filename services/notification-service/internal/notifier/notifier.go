package notifier

import (
	"fmt"
	"log"
	"time"
)

// Notifier เป็น interface เผื่อเปลี่ยนวิธีแจ้งเตือน (Email/LINE/Slack/SMS)
type Notifier interface {
	Notify(subject, message string) error
}

// ConsoleNotifier (MVP) — log ออก console
type ConsoleNotifier struct{}

func NewConsole() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (c *ConsoleNotifier) Notify(subject, message string) error {
	log.Printf("[notify] %s :: %s\n", subject, message)
	return nil
}

// Helper ทำให้ข้อความอ่านง่าย
func HumanTimeRange(startUnix, endUnix int64) string {
	st := time.Unix(startUnix, 0).Local()
	et := time.Unix(endUnix, 0).Local()
	return fmt.Sprintf("%s — %s", st.Format("2006-01-02 15:04"), et.Format("15:04"))
}
