package main

import (
    "log"
    "bytes"
    "golang.org/x/crypto/ssh"
    "fmt"
    "io/ioutil"
    "os"
    "time"
    "net"
    "golang.org/x/crypto/ssh/agent"
)

func makeSigner(keyname string) (signer ssh.Signer, err error) {
    fp, err := os.Open(keyname)
    if err != nil {
        return
    }
    defer fp.Close()

    buf, _ := ioutil.ReadAll(fp)
    signer, _ = ssh.ParsePrivateKey(buf)
    return
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
    conn, err := ssh.Dial("tcp", hostname+":22", config)
    if err != nil {
          log.Fatal("Failed to dial: ", err)
    }

    session, err := conn.NewSession()
    if err != nil {
          log.Fatal("Failed to create session: ", err)
     }
    defer session.Close()

		modes := ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}

		if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
			session.Close()
			fmt.Errorf("request for pseudo terminal failed: %s", err)
		}

    var stdoutBuf bytes.Buffer
    session.Stdout = &stdoutBuf
    session.Run(cmd)

    return "########### " + hostname + " ########## \n" +
            stdoutBuf.String() +
            "\n##################################################################################\n\n"
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func main() {
    cmd := "echo \"### HOSTNAME $(hostname) ###\n\n\";" + os.Args[1] // In case we use ip address instead of FQDN
    hosts := os.Args[2:]

    results := make(chan string, 10)
    timeout := time.After(120 * time.Second)
    config := &ssh.ClientConfig{
        User: os.Getenv("LOGNAME"),
        Auth: []ssh.AuthMethod{SSHAgent()},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    for _, hostname := range hosts {
        go func(hostname string) {
            results <- executeCmd(cmd, hostname, config)
        }(hostname)
    }

    for i := 0; i < len(hosts); i++ {
        select {
        case res := <-results:
            fmt.Print(res)
        case <-timeout:
            fmt.Println("Timed out!")
            return
        }
    }
}
