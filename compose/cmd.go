package compose

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"
	"time"
)

var AlreadyDoneErr = errors.New("cmd already done")

type Cmd struct {
	Config *CmdConfig

	mu          sync.Mutex
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	readyRegexp *regexp.Regexp
	ready       bool
	readyChan   chan struct{}
	running     bool
	doneChan    chan struct{}
	cmd         *exec.Cmd
}

func NewCmd(config *CmdConfig) *Cmd {
	cmd := &Cmd{
		Config:    config,
		readyChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
	}

	if config.ReadyWhenLog != "" {
		cmd.readyRegexp = regexp.MustCompile(config.ReadyWhenLog)
	}

	return cmd
}

func (c *Cmd) Start() error {
	log.Printf("%v cmd.Start: %v\n", c.Config.Name, c.Config.Cmd)
	cmd := exec.Command(c.Config.Cmd[0], c.Config.Cmd[1:]...)
	// don't send ctrl+c signal to child process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	c.stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	c.stderr = stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	c.cmd = cmd
	c.running = true
	go c.run()
	return nil
}

func (c *Cmd) Ready() error {
	<-c.readyChan

	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return AlreadyDoneErr
	}
	return nil
}

func (c *Cmd) Wait() {
	<-c.doneChan
}

func (c *Cmd) Kill() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	log.Printf("%v cmd.Kill\n", c.Config.Name)
	if err := c.cmd.Process.Kill(); err != nil {
		log.Printf("%v process.Kill error: %v\n", c.Config.Name, err)
	}
}

func (c *Cmd) done() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}
	c.running = false
	close(c.doneChan)
}

func (c *Cmd) switchToReady() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ready {
		return
	}
	c.ready = true
	close(c.readyChan)
}

func (c *Cmd) watchReadyTimeout() {
	timeout := c.Config.ReadyTimeout
	if timeout <= 0 {
		timeout = DefaultReadyTimeout
	}

	select {
	case <-time.After(timeout):
		c.switchToReady()
	case <-c.readyChan:
	}
}

func (c *Cmd) watchReadyLine(line string) {
	if c.ready {
		return
	}
	if c.readyRegexp == nil {
		return
	}

	if c.readyRegexp.MatchString(line) {
		c.switchToReady()
	}
}

func (c *Cmd) pipe(dst io.Writer, src io.Reader) {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		c.watchReadyLine(line)
		fmt.Fprintf(dst, "%s | %s\n", c.Config.Name, line)
	}
}

func (c *Cmd) run() {
	defer c.done()
	defer c.switchToReady()

	go c.watchReadyTimeout()
	go c.pipe(os.Stdout, c.stdout)
	go c.pipe(os.Stderr, c.stderr)

	if err := c.cmd.Wait(); err != nil {
		log.Printf("%v cmd.Wait error: %v\n", c.Config.Name, err)
	}
}
