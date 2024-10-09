package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"line_counter/config"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

type User struct {
	Name          string
	CodeLineCount int
}

// Getting data on the number of rows that users pushed into git
func main() {
	userLineCount := make(map[string][]User)

	userLineCount = processProjects(config.ProjectsPaths, userLineCount)

	userLineCount = sortMapByCodeLines(userLineCount)

	printByProjects(userLineCount)

	printAllProjects(userLineCount)
}

func processProjects(projectsPaths []string, usersLines map[string][]User) map[string][]User {
	for _, path := range projectsPaths {
		_ = os.Chdir(path)

		usersLines[path] = getCodeLinesFromFile(path, usersLines[path])
	}

	return usersLines
}

func getCodeLinesFromFile(directory string, userCodeLinesCount []User) []User {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Printf("Error reading directory: %s, %v\n", directory, err)
		return userCodeLinesCount
	}

	for _, file := range files {
		fullFilePath := directory + "/" + file.Name()

		if isNameForbidden(file.Name()) {
			continue
		}

		if file.IsDir() {
			userCodeLinesCount = getCodeLinesFromFile(fullFilePath, userCodeLinesCount)
		} else if isFileHasApprovedExtension(file.Name()) {
			gitBlameOutput, err := gitBlameCommand(fullFilePath)

			if err != nil {
				fmt.Println("Error getting git blame output", err)
			}

			userCodeLinesCount = countLinesPerUser(gitBlameOutput, userCodeLinesCount)
		}
	}

	return userCodeLinesCount
}

func isNameForbidden(fileName string) bool {
	for _, forbiddenFolderName := range config.ForbiddenFileAndFolderNames {
		if fileName == forbiddenFolderName {
			return true
		}
	}

	return false
}

func isFileHasApprovedExtension(fileName string) bool {
	for _, extension := range config.ApprovedExtensions {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}

	return false
}

func gitBlameCommand(filePath string) (string, error) {
	cmd := exec.Command("git", "blame", filePath)

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running git blame: %s", stderr.String())
	}

	return out.String(), nil
}

func countLinesPerUser(gitBlameOutput string, userCodeLinesCount []User) []User {
	re := regexp.MustCompile(`^\^?\w+\s+\((.*?)\s+\d{4}-\d{2}-\d{2}.*?\)\s+(.*)$`)

	scanner := bufio.NewScanner(strings.NewReader(gitBlameOutput))
	for scanner.Scan() {
		line := scanner.Text()

		matches := re.FindStringSubmatch(line)
		if len(matches) > 2 {
			author := matches[1]
			content := matches[2]

			if content != "" {
				authorHasFound := false

				for index, user := range userCodeLinesCount {
					if user.Name == author {
						userCodeLinesCount[index].CodeLineCount++
						authorHasFound = true
						break
					}
				}

				if !authorHasFound {
					userCodeLinesCount = append(userCodeLinesCount, User{
						Name:          author,
						CodeLineCount: 1,
					})
				}
			}
		}
	}

	return userCodeLinesCount
}

func sortMapByCodeLines(userLineCount map[string][]User) map[string][]User {
	for project, users := range userLineCount {
		userLineCount[project] = sortUserArrayByCodeLines(users)
	}

	return userLineCount
}

func sortUserArrayByCodeLines(usersLineCounts []User) []User {
	sort.Slice(usersLineCounts, func(i, j int) bool {
		return usersLineCounts[i].CodeLineCount > usersLineCounts[j].CodeLineCount
	})

	return usersLineCounts
}

func printByProjects(userLineCount map[string][]User) {
	for path, users := range userLineCount {
		printProjectTitle(path)

		printAllUsers(users)

		fmt.Println()
	}
}

func printAllProjects(userLineCount map[string][]User) {
	allUsersCodeLines := getAllProjectsUsersCodeLine(userLineCount)

	allUsersCodeLines = sortUserArrayByCodeLines(allUsersCodeLines)

	printProjectTitle("Total")

	printAllUsers(allUsersCodeLines)
}

func getAllProjectsUsersCodeLine(usersLinesCount map[string][]User) []User {
	userTotalLines := make(map[string]int)

	for _, users := range usersLinesCount {
		for _, user := range users {
			userTotalLines[user.Name] += user.CodeLineCount
		}
	}

	var allUsersCodeLines []User
	for name, lineCount := range userTotalLines {
		allUsersCodeLines = append(allUsersCodeLines, User{Name: name, CodeLineCount: lineCount})
	}

	return allUsersCodeLines
}

func printAllUsers(UsersLinesCode []User) {
	for _, user := range UsersLinesCode {
		fmt.Printf("%s: %d lines\n", user.Name, user.CodeLineCount)
	}
}

func printProjectTitle(projectName string) {
	fmt.Println("---------------------")
	fmt.Printf("Project: %s\n", projectName)
	fmt.Println("---------------------")
}
