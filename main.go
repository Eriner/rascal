package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"go.etcd.io/bbolt"
)

const (
	maxConcurrency = 1024
)

type Bot struct {
	ircevent.Connection
	owner     string
	userAgent string
	semaphore chan struct{}
	db        *bbolt.DB
}

func (b *Bot) acquireSemaphore() bool {
	select {
	case b.semaphore <- struct{}{}:
		return true
	default:
		return false
	}
}

func (b *Bot) releaseSemaphore() {
	<-b.semaphore
}

func (b *Bot) sendReplyNotice(target, msgid, text string) {
	if len(msgid) == 0 {
		b.Notice(target, text)
		return
	}
	b.SendWithTags(map[string]string{"+draft/reply": msgid}, "NOTICE", target, text)
}

func isFromOwner(e ircmsg.Message, owner string) bool {
	if len(owner) == 0 {
		return false
	}
	if present, account := e.GetTag("account"); present && account == owner {
		return true
	}
	return false
}

func newBot() *Bot {
	nick := os.Getenv("RASCAL_NICK")
	server := os.Getenv("RASCAL_SERVER")
	channels := os.Getenv("RASCAL_CHANNELS")
	saslLogin := os.Getenv("RASCAL_SASL_LOGIN")
	saslPassword := os.Getenv("RASCAL_SASL_PASSWORD")
	owner := os.Getenv("RASCAL_OWNER")
	version := os.Getenv("RASCAL_VERSION")
	quitMessage := os.Getenv("RASCAL_QUIT_MESSAGE")
	if len(version) == 0 {
		version = "rascal (dev)"
	}
	debug := len(os.Getenv("RASCAL_DEBUG")) > 0
	userAgent := os.Getenv("RASCAL_USER_AGENT")
	dbPath := os.Getenv("RASCAL_DB_PATH")
	if len(dbPath) == 0 {
		dbPath = "rascal.db"
	}

	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("karma"))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		_ = b
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	irc := &Bot{
		Connection: ircevent.Connection{
			Server:       server,
			Nick:         nick,
			UseTLS:       true,
			TLSConfig:    &tls.Config{InsecureSkipVerify: true},
			RequestCaps:  []string{"server-time", "message-tags", "account-tag", "draft/relaymsg=/"},
			SASLLogin:    saslLogin,
			SASLPassword: saslPassword,
			Version:      version,
			QuitMessage:  quitMessage,
			Debug:        debug,
		},
		owner:     owner,
		userAgent: userAgent,
		semaphore: make(chan struct{}, maxConcurrency),
		db:        db,
	}

	irc.AddConnectCallback(func(m ircmsg.Message) {
		if botMode := irc.ISupport()["BOT"]; len(botMode) > 0 {
			irc.Send("MODE", irc.CurrentNick(), "+"+botMode)
		}
		for _, ch := range strings.Split(channels, ",") {
			irc.Join(strings.TrimSpace(ch))
		}
	})
	irc.AddCallback("PRIVMSG", func(m ircmsg.Message) {
		target, msg := m.Params[0], m.Params[1]
		_, msgid := m.GetTag("msgid")
		if strings.HasPrefix(msg, irc.Nick) {
			if !isFromOwner(m, irc.owner) {
				irc.sendReplyNotice(m.Nick(), msgid, "no, stalker child")
				return
			}
		}
		if strings.Contains(msg, "relay: ") {
			if isFromOwner(m, irc.owner) {
				msg := strings.Split(msg, "relay: ")[1]
				irc.Send("RELAYMSG", target, "eriner/NAS", msg)
			}
		}
		if strings.Contains(msg, "++") {
			user := strings.Split(msg, "++")[0]
			idx := strings.LastIndex(user, " ")
			if idx > -1 {
				user = user[idx+1:]
			}
			// TODO either check against list of nicks (complicated) or regexp user
			karma := 0
			if err := irc.db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte("karma"))
				u := []byte(user)
				var err error
				ck := string(b.Get(u))
				if len(ck) == 0 {
					ck = "0"
				}
				karma, err = strconv.Atoi(ck)
				if err != nil {
					return fmt.Errorf("get karma: %w", err)
				}
				karma++
				if err := b.Put(u, []byte(strconv.Itoa(karma))); err != nil {
					return fmt.Errorf("put karma: %w", err)
				}
				return nil
			}); err != nil {
				log.Println("error adding karma: %w", err)
				return
			}
			irc.Actionf(target, "%s gained +1 karma. Total karma: %d", user, karma)
			return
		}
	})
	irc.AddCallback("INVITE", func(m ircmsg.Message) {
		irc.Join(m.Params[1])
	})
	return irc
}

func main() {
	bot := newBot()
	defer bot.db.Close()
	if err := bot.Connect(); err != nil {
		log.Fatal(err)
	}
	bot.Loop()
}
