package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type JiraCredentials struct {
	Host     string
	Username string
	Token    string
}

func main() {
	credentials, err := getCredentials()
	handleError(err)

	issueTag, err := getCurrentIssueTag()
	handleError(err)

	tp := jira.BasicAuthTransport{
		Username: credentials.Username,
		Password: credentials.Token,
	}

	client, err := jira.NewClient(tp.Client(), credentials.Host)
	handleError(err)

	var time string
	os.Stderr.Write([]byte("Please enter worktime: "))
	fmt.Scanln(&time)

	_, _, err = client.Issue.AddWorklogRecord(issueTag, &jira.WorklogRecord{
		TimeSpent: time,
	})

	if err != nil {
		handleError(fmt.Errorf("failed to add time to %s. Error: %s", issueTag, err))
	}

	fmt.Printf("Logged '%s' to issue '%s'", "1m", issueTag)
}

func handleError(err error) {
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
}

func getCredentials() (JiraCredentials, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return JiraCredentials{}, fmt.Errorf("failed to get user directory. Error: %s", err.Error())
	}

	fileContents, err := os.ReadFile(fmt.Sprintf("%s/.jiratt", homeDir))
	if err != nil {
		return JiraCredentials{}, err
	}

	credentials := JiraCredentials{}
	credentialParts := strings.Split(strings.Trim(string(fileContents), "\n"), " ")
	credentialFlags := map[string]bool{"host": false, "user": false, "token": false}
	for i, content := range credentialParts {
		if i%2 == 1 {
			continue
		}

		switch content {
		case "host":
			credentials.Host = credentialParts[i+1]
			credentialFlags["host"] = true
			break
		case "user":
			credentials.Username = credentialParts[i+1]
			credentialFlags["user"] = true
			break
		case "token":
			credentials.Token = credentialParts[i+1]
			credentialFlags["token"] = true
		default:
			return JiraCredentials{}, fmt.Errorf("unknown credentials property '%s'", content)
		}
	}

	for flagKey, flag := range credentialFlags {
		if flag == false {
			return JiraCredentials{}, fmt.Errorf("credentials property '%s' is not set", flagKey)
		}
	}

	return credentials, nil
}

func getCurrentIssueTag() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	branch, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get issue tag from branch. Error: %s", err.Error())
	}

	if err != nil {
		return "", fmt.Errorf("failed to get issue tag from branch. Error: %s", err.Error())
	}

	branchRxp := regexp.MustCompile("^(FNX-\\d+)")
	branchMatches := branchRxp.FindStringSubmatch(string(branch))
	if len(branchMatches) == 0 {
		return "", fmt.Errorf("branch does not follow the 'TAG-xxx' convention")
	}

	return branchMatches[0], nil
}
