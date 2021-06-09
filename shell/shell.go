package shell

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

// Shell reperesents a readable and writable slave pty.
// Writing to file is equal to typing in a terminal
// and Reading from file receives output that should
// normally go to master pty.
type Shell struct {
	file    *os.File
	output  chan string
	stopped chan bool
}

// New returns instance of Shell. If New
// returned no error the shell is fully functional
// (readable and writable when New returns.
// onstartscript is the name of the file sourced
// when shell starts.
func New(onstartscript string) (*Shell, error) {
	f, err := startShell()
	if err != nil {
		return nil, err
	}
	var sh Shell
	sh.file = f
	sh.output = make(chan string, 1)
	sh.stopped = make(chan bool, 1)
	if err := sh.executeNoOut("stty -echo"); err != nil {
		// failure at this point means sh.file is not writable
		return nil, err
	}
	// echo is guaranteed to be off from this point on
	// and sh.file is writable and readable
	go func() {
		if onstartscript != "" {
			sh.output <- "sourcing " + onstartscript
			sh.Execute("source " + onstartscript)
		}
		buf := make([]byte, 16384)
		for {
			// This call to Read() unblocks when shell quits on 'exit' command
			// otherwise this goroutine gets stuck and garbage collected.
			n, err := sh.file.Read(buf)
			if err != nil {
				// log.Println("shell: reader goroutine stopping on error:", err)
				break
			}
			sh.output <- string(buf[:n])
		}
		close(sh.output)
		close(sh.stopped)
		// log.Println("shell: reader goroutine finished")
	}()
	return &sh, err
}

// startShell starts bash and returns *os.File
// writing to file == writing to bash stdin
// reading from file == reading from bash stdout
func startShell() (*os.File, error) {
	// trying to convince everyone we need compatibility mode
	os.Setenv("TERM", "vt220")
	cmd := exec.Command("bash")
	pts, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return pts, nil
}

// Output returns a channel to which the slave pty writes
func (sh *Shell) Output() chan string {
	return sh.output
}

// Stopped returns a channel indicating that underlying shell died.
// If read from this channel succeeds  no more output will be sent
// to channel accesible with Output() method.
// Either the shell quitted on 'exit' command, due to an io error.
func (sh *Shell) Stopped() chan bool {
	return sh.stopped
}

// Execute sends the command to underlying shell
func (sh *Shell) Execute(command string) error {
	payload := []byte(command + "\n")
	if err := sh.writeCommand(payload); err != nil {
		return fmt.Errorf("error running '%v': %v", command, err)
	}
	return nil
}

// this is intended to send command to shell and suppress output
// by waiting and reading what shell has to offer.
func (sh *Shell) executeNoOut(command string) error {
	if err := sh.Execute(command); err != nil {
		return err
	}
	// assuming output of shell that we're intending to flush
	// will not exceed 4096 bytes
	wasteBuf := make([]byte, 4096)
	// wait for the shell to have output ready
	// if we read immediately - there is a chance that
	// shell will output somethin after Read returns
	time.Sleep(time.Second)
	sh.file.Read(wasteBuf)
	return nil
}

// writeCommand does the actual write to the shell input
func (sh *Shell) writeCommand(payload []byte) error {
	if _, err := sh.file.Write(payload); err != nil {
		return fmt.Errorf("error writing to pty: %w", err)
	}
	return nil
}

// Stop is intended to shut down the shell gracefully
// without sending 'exit' command to shell's input.
// Not to be used for the moment since in current form it just
// leaves orphan goroutine from New func above hanging forever.
// When the goroutine blocks on Read closing the file (in this method)
// does not unblock Read. TODO.
func (sh *Shell) Stop() {
	// TODO: better way to quit
	if err := sh.file.Close(); err != nil {
		log.Println("shell: file.Close() returned:", err)
	}
	sh.stopped <- true
}
