package notifier

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/trancee/DealScout/internal/deal"
)

// Notifier sends deal notifications to Telegram.
type Notifier struct {
	botToken   string
	channelID  string
	topics     map[string]int
	apiBase    string
	lastSendAt time.Time
}

// New creates a Telegram Notifier.
func New(botToken, channelID string, topics map[string]int) *Notifier {
	return &Notifier{
		botToken:  botToken,
		channelID: channelID,
		topics:    topics,
		apiBase:   "https://api.telegram.org",
	}
}

// WithAPIBase overrides the Telegram API base URL (for testing).
func (n *Notifier) WithAPIBase(base string) *Notifier {
	n.apiBase = base
	return n
}

// Send posts a deal notification to the appropriate Telegram topic.
func (n *Notifier) Send(d deal.Deal) error {
	n.rateLimit()

	caption := FormatCaption(d)
	topicID := n.topics[d.Category]

	if d.ImageURL != "" {
		if err := n.sendPhoto(d.ImageURL, caption, topicID); err == nil {
			return nil
		}
		slog.Warn("sendPhoto failed, falling back to sendMessage", "product", d.ProductName)
	}

	return n.sendMessage(caption, topicID)
}

func (n *Notifier) sendPhoto(imageURL, caption string, topicID int) error {
	params := url.Values{
		"chat_id":           {n.channelID},
		"photo":             {imageURL},
		"caption":           {caption},
		"parse_mode":        {"MarkdownV2"},
		"message_thread_id": {fmt.Sprintf("%d", topicID)},
	}

	apiURL := fmt.Sprintf("%s/bot%s/sendPhoto", n.apiBase, n.botToken)
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendPhoto returned %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) sendMessage(text string, topicID int) error {
	params := url.Values{
		"chat_id":           {n.channelID},
		"text":              {text},
		"parse_mode":        {"MarkdownV2"},
		"message_thread_id": {fmt.Sprintf("%d", topicID)},
	}

	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", n.apiBase, n.botToken)
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendMessage returned %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) rateLimit() {
	elapsed := time.Since(n.lastSendAt)
	if elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}
	n.lastSendAt = time.Now()
}

// FormatCaption builds a MarkdownV2 caption for a deal notification.
func FormatCaption(d deal.Deal) string {
	name := EscapeMarkdownV2(d.ProductName)
	price := EscapeMarkdownV2(fmt.Sprintf("CHF %.2f", d.Price))
	shop := EscapeMarkdownV2(d.Shop)

	var lines []string
	lines = append(lines, fmt.Sprintf("🔥 %s", name))
	lines = append(lines, fmt.Sprintf("💰 %s", price))

	if !d.IsFirstSeen && d.DiscountPct > 0 {
		drop := EscapeMarkdownV2(fmt.Sprintf("-%.0f%% (was CHF %.2f)", d.DiscountPct, d.LastDBPrice))
		lines = append(lines, fmt.Sprintf("📉 %s", drop))
	}

	if d.OldPrice != nil {
		shopPrice := EscapeMarkdownV2(fmt.Sprintf("CHF %.2f", *d.OldPrice))
		lines = append(lines, fmt.Sprintf("🏷️ Shop says: %s", shopPrice))
	}

	lines = append(lines, fmt.Sprintf("🏪 %s", shop))

	if d.URL != "" {
		escapedURL := EscapeMarkdownV2(d.URL)
		lines = append(lines, fmt.Sprintf("🔗 [View Product](%s)", escapedURL))
	}

	return strings.Join(lines, "\n")
}

// EscapeMarkdownV2 escapes special characters for Telegram MarkdownV2.
func EscapeMarkdownV2(s string) string {
	replacer := strings.NewReplacer(
		"_", `\_`,
		"*", `\*`,
		"[", `\[`,
		"]", `\]`,
		"(", `\(`,
		")", `\)`,
		"~", `\~`,
		"`", "\\`",
		">", `\>`,
		"#", `\#`,
		"+", `\+`,
		"-", `\-`,
		"=", `\=`,
		"|", `\|`,
		"{", `\{`,
		"}", `\}`,
		".", `\.`,
		"!", `\!`,
	)
	return replacer.Replace(s)
}
