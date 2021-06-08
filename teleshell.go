package teleshell

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/dmfed/teleshell/shell"
	"gopkg.in/tucnak/telebot.v2"
)

// 4096 is Telegram's max message size.
const maxTelegramMessageSize = 4096

var (
	botCommandRegexp = regexp.MustCompile(`^/(\w+) *`)
)

var (
	commStartSession  = "shell"
	commStopSession   = "exit"
	commSingleCommand = "cmd"
	commHelp          = "help"
)

var (
	msgSessionStarted     = "Shell started. You can talk to your machine now. Say 'exit' to stop the  shell."
	msgErrStartingSession = "Could not start shell."
	msgSessionStopped     = "Shell stopped. You are no longer talking to your machine."
	// msgErrStoppingSession   = "Could not stop shell."
	msgSessionInProgress    = "Your shell session is in progress. Say 'exit' to stop it."
	msgSessionNotInProgress = "There are no active shell sessions."
	msgNotAuthorized        = "Sorry, you are not permitted to issue commands."
)

var msgHelp string = fmt.Sprintf(`Welcome to teleshell!
Use the following commands:
/%v <command> to run a single command on your machine 
	without launching shell.
/%v to start bash shell on your machine and redirect 
	input from this chat to the shell. Avoid launching 
	interactive programs. sudo is OK, but vim is NOT. 
	Also colored output of programs appears as 
	garbage in chat.
/%v to force-kill running shell
/%v to see this message again.`, commSingleCommand, commStartSession, commStopSession, commHelp)

type TeleShell struct {
	TelegramUsername    string
	OnStart             string
	shellChatInProgress bool
	shell               *shell.Shell
	bot                 *telebot.Bot
}

// New returns instance of TeleShell with default settings.
// Call Start() method (blocking) to make the bot accept messages.
func New(token, username, onstartscript string) (*TeleShell, error) {
	// init the bot.
	// These are default settings from example code in
	// telebot docs
	settings := telebot.Settings{Token: token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second}}
	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}
	// init state
	var ts TeleShell
	ts.TelegramUsername = username
	ts.OnStart = onstartscript
	ts.bot = bot
	// single handler since we only want to talk to authorized user
	ts.bot.Handle(telebot.OnText, ts.authAndRoute)
	return &ts, nil
}

// Start starts Teleshell and embedded bot. Calling Start is blocking.
func (ts *TeleShell) Start() {
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		sig := <-interrupts
		log.Printf("teleshell exiting on signal: %v", sig)
		ts.Stop()
	}()
	ts.bot.Start()
}

// Stop stops TeleShell and embedded bot.
func (ts *TeleShell) Stop() {
	if ts.shell != nil {
		ts.shell.Stop()
	}
	ts.bot.Stop()
}

func (ts *TeleShell) authAndRoute(m *telebot.Message) {
	if !ts.isAuthorized(m.Sender) {
		ts.sorry(m)
		return
	}
	command := ""
	if botCommandRegexp.MatchString(m.Text) {
		command = botCommandRegexp.FindStringSubmatch(m.Text)[1]
	}
	switch {
	case command == commSingleCommand:
		ts.execSingleCommand(m)
	case command == commStartSession:
		ts.startSession(m)
	case command == commStopSession:
		ts.stopSession(m)
	case command == commHelp:
		ts.help(m)
	case ts.shellChatInProgress:
		ts.handleSessionMsg(m)
	default:
		ts.help(m)
	}
}

func (ts *TeleShell) isAuthorized(u *telebot.User) bool {
	return u.Username == ts.TelegramUsername && !u.IsBot
}

func (ts *TeleShell) execSingleCommand(m *telebot.Message) {
	command := stripBotCommand(m.Text)
	output, err := execCmd(command)
	ts.send(m.Chat, output)
	if err != nil {
		ts.send(m.Chat, err.Error())
	}
}

func (ts *TeleShell) startSession(m *telebot.Message) {
	if ts.shellChatInProgress {
		ts.send(m.Chat, msgSessionInProgress)
		return
	}
	shell, err := shell.New(ts.OnStart)
	if err != nil {
		ts.send(m.Chat, msgErrStartingSession)
		return
	}
	ts.shell = shell
	ts.shellChatInProgress = true
	ts.send(m.Chat, msgSessionStarted)
	go func() {
		for ts.shellChatInProgress { // TODO: data race here
			select {
			case <-shell.Stopped():
				ts.shellChatInProgress = false
			case output := <-shell.Output():
				ts.send(m.Chat, output)
			}
		}
		ts.send(m.Chat, msgSessionStopped)
		// log.Println("teleshell: shell listening goroutine finishing")
	}()
}

func (ts *TeleShell) stopSession(m *telebot.Message) {
	if !ts.shellChatInProgress {
		ts.send(m.Chat, msgSessionNotInProgress)
		return
	}
	ts.shellChatInProgress = false // this will prevent wait on select
	ts.shell.Stop()                // this will trigger release from select
	time.Sleep(time.Second)
	ts.shell = nil
}

func (ts *TeleShell) handleSessionMsg(m *telebot.Message) {
	if err := ts.shell.Execute(m.Text); err != nil {
		ts.send(m.Chat, err.Error())
	}
}

func (ts *TeleShell) send(c *telebot.Chat, msg string) {
	if len(msg) > maxTelegramMessageSize {
		ts.paginatedSend(c, msg)
	} else {
		ts.bot.Send(c, msg)
	}
}

func (ts *TeleShell) paginatedSend(c *telebot.Chat, msg string) {
	messages := paginate(msg)
	for _, message := range messages {
		ts.bot.Send(c, message)
	}
}

func (ts *TeleShell) help(m *telebot.Message) {
	ts.send(m.Chat, msgHelp)
}

func (ts *TeleShell) sorry(m *telebot.Message) {
	ts.send(m.Chat, msgNotAuthorized)
}

func (ts *TeleShell) inProgress(m *telebot.Message) {
	ts.send(m.Chat, msgSessionInProgress)
}

func stripBotCommand(message string) string {
	loc := botCommandRegexp.FindStringIndex(message)
	return message[loc[1]:]
}

func execCmd(c string) (string, error) {
	commSlice := strings.Split(c, " ")
	var cmd *exec.Cmd
	if len(commSlice) > 1 {
		cmd = exec.Command(commSlice[0], commSlice[1:]...)
	} else {
		cmd = exec.Command(commSlice[0])
	}
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}

func paginate(input string) []string {
	pages := []string{}
	buf := bytes.Buffer{}
	// make sure we do not split a single line of text
	for _, s := range strings.Split(input, "\n") {
		// assuming single line of input (s) can not exceed maxTelegramMessageSize bytes
		if buf.Len()+len(s) > maxTelegramMessageSize {
			pages = append(pages, buf.String())
			buf.Reset()
		}
		buf.WriteString(s + "\n")
	}
	pages = append(pages, buf.String())
	return pages
}
