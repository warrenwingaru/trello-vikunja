package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/warrenwingaru/go-trello"
	"github.com/yuin/goldmark"
	"os"
	"runtime"
	"strconv"
	"strings"
	"wingaru.me/trello-migrate/internal/migration"
	"wingaru.me/trello-migrate/internal/models"
	"wingaru.me/trello-migrate/pkg/vikunja"
)

var trelloColorMap map[string]string
var trelloApiKey string
var trelloApiToken string
var vikunjaApiKey string

const maxTaskSize = 200

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(".env file not found")
	}
	trelloApiKey = os.Getenv("TRELLO_API_KEY")
	trelloApiToken = os.Getenv("TRELLO_API_TOKEN")
	vikunjaApiKey = os.Getenv("VIKUNJA_API_KEY")

	trelloColorMap = make(map[string]string, 30)
	trelloColorMap = map[string]string{
		"green":        "4bce97",
		"yellow":       "f5cd47",
		"orange":       "fea362",
		"red":          "f87168",
		"purple":       "9f8fef",
		"blue":         "579dff",
		"sky":          "6cc3e0",
		"lime":         "94c748",
		"pink":         "e774bb",
		"black":        "8590a2",
		"green_dark":   "1f845a",
		"yellow_dark":  "946f00",
		"orange_dark":  "c25100",
		"red_dark":     "c9372c",
		"purple_dark":  "6e5dc6",
		"blue_dark":    "0c66e4",
		"sky_dark":     "227d9b",
		"lime_dark":    "5b7f24",
		"pink_dark":    "ae4787",
		"black_dark":   "626f86",
		"green_light":  "baf3db",
		"yellow_light": "f8e6a0",
		"orange_light": "fedec8",
		"red_light":    "ffd5d2",
		"purple_light": "dfd8fd",
		"blue_light":   "cce0ff",
		"sky_light":    "c6edfb",
		"lime_light":   "d3f1a7",
		"ping_light":   "fdd0ec",
		"black_light":  "dcdfe4",
		"transparent":  "", // Empty
	}

}

func getPadding(padding int) string {
	if padding <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", padding, "")
}

var boardsToMigrate []string

func readInputUnix() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the numbers of board to migrate (1, 3, 4): ")
	text, _ := reader.ReadString('\n')
	return text
}

func readInputWindows() string {
	fmt.Print("Enter the numbers of board to migrate (1, 3, 4): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()
	return text
}

func main() {
	Init()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	client := vikunja.NewClient(vikunjaApiKey, os.Getenv("VIKUNJA_INSTANCE"))
	client.Logger = logger
	vikunjaData, err := readDataFile("data.json")
	if err != nil {
		panic(err)
	}

	boardNames := make([]string, 0, len(vikunjaData))

	for name := range vikunjaData {
		boardNames = append(boardNames, name)
	}

	// output options
	maxLength := 0
	for _, option := range boardNames {
		if len(option) > maxLength {
			maxLength = len(option)
		}
	}

	for i, option := range boardNames {
		padding := maxLength - len(option) + 3
		if i > 8 { // For two-digit indices
			padding--
		}
		if i%2 == 0 {
			fmt.Printf("%d) %s%s", i+1, option, getPadding(padding))
		} else {
			fmt.Printf("%d) %s%s\n", i+1, option, getPadding(padding))
		}
	}

	var text string
	if strings.Contains(strings.ToLower(runtime.GOOS), "windows") {
		text = readInputWindows()
	} else {
		text = readInputUnix()
	}
	chosenInStr := strings.Split(strings.TrimSpace(text), ",")
	chosen := make([]int, 0, len(chosenInStr))
	for _, str := range chosenInStr {
		num, err := strconv.Atoi(str)
		if err != nil {
			fmt.Printf("Error converting string to integer: %v\n\n", err)
			return
		}
		chosen = append(chosen, num)
	}

	boardsToMigrate = make([]string, 0, len(chosen))
	for _, option := range chosen {
		boardsToMigrate = append(boardsToMigrate, boardNames[option-1])
	}

	trelloData, err := readTrelloFile("trello.json")
	if err != nil {
		panic(err)
	}

	data, err := convertTrelloToVikunja(trelloData, vikunjaData)

	// upload
	for _, board := range data {
		// dry run
		if len(board.Buckets) > 0 {
			logger.Debugf("Uploading %d bucket", len(board.Buckets))
		}
		for _, bucket := range board.Buckets {
			// create a bucket
			err := client.CreateBucket(bucket)
			if err != nil {
				panic(err)
			}

			if len(bucket.TasksWithComments) > 0 {
				logger.Debugf("Uploading %d tasks", len(bucket.TasksWithComments))
			}

			for _, task := range bucket.TasksWithComments {
				newTask := &task.Task
				newTask.BucketID = bucket.ID

				err := client.AddTask(newTask)
				if err != nil {
					panic(err)
				}

				if len(task.Comments) > 0 {
					logger.Debugf("Uploading %d comments", len(task.Comments))
				}
				// add task comments
				for _, comment := range task.Comments {
					comment.TaskID = newTask.ID
					err := client.AddTaskComment(comment)
					if err != nil {
						panic(err)
					}
				}

				if len(task.Attachments) > 0 {
					logger.Debugf("Uploading %d attachments", len(task.Attachments))
				}
				for _, attachment := range task.Attachments {
					if len(attachment.File.FileContent) > 0 {
						err = client.AddTaskAttachments(newTask.ID, attachment)
						if err != nil {
							panic(err)
						}
					}
				}
			}

		}

	}
}

func isBoardInList(item string, array []string) bool {
	for _, value := range array {
		if item == value {
			return true
		}
	}

	return false

}

func readTrelloFile(filename string) ([]*trello.Board, error) {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var result []*trello.Board

	if err := decoder.Decode(&result); err != nil {
		fmt.Println("Error decoding file:", err)
		return nil, err
	}
	return result, nil
}

func readDataFile(filename string) (map[string]models.Project, error) {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var result []models.Project

	if err := decoder.Decode(&result); err != nil {
		fmt.Println("Error decoding file:", err)
		return nil, err
	}
	dataMap := make(map[string]models.Project, len(result))
	for _, data := range result {
		dataMap[data.Title] = data
	}

	return dataMap, nil
}

func convertTrelloToVikunja(boards []*trello.Board, vikunjaData map[string]models.Project) (hierarchy []*models.ProjectWithTasksAndBuckets, err error) {
	fmt.Printf("[Trello Migration] Converting %d boards to vikunja projects\n", len(boards))

	for _, board := range boards {
		if projectFromData, found := vikunjaData[board.Name]; found {
			if !isBoardInList(board.Name, boardsToMigrate) {
				continue
			}
			project := &models.ProjectWithTasksAndBuckets{
				Project: models.Project{
					ID:          projectFromData.ID,
					Title:       board.Name,
					Description: board.Desc,
				},
			}
			// create bucket for each view or maybe for kanban only
			for _, view := range projectFromData.Views {
				if view.Title == "Kanban" {
					var buckets []*models.Bucket
					buckets = []*models.Bucket{}
					var tasks []*models.TaskWithComments
					tasks = []*models.TaskWithComments{}

					// Create tasks with the new bucket
					for _, l := range board.Lists {

						fmt.Printf("[Trello Migration] Converting %d cards to tasks from board %s\n", len(l.Cards), board.Name)
						for _, card := range l.Cards {
							fmt.Printf("[Trello Migration] Conveting card %s\n", card.Name)

							task := &models.TaskWithComments{
								Task: models.Task{
									Title:     card.Name,
									ProjectID: projectFromData.ID,
								},
							}

							task.Description, _ = convertMarkdownToHTML(card.Desc)

							for _, checklist := range card.Checklists {
								task.Description += "\n\n<h2> " + checklist.Name + "</h2>\n\n" + `<ul data-type="taskList">`
								for _, item := range checklist.CheckItems {
									task.Description += "\n"
									if item.State == "complete" {
										task.Description += `<li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>` + item.Name + `</p></div></li>`
									} else {
										task.Description += `<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>` + item.Name + `</p></div></li>`
									}
								}
								task.Description += "</ul>"
							}
							if len(card.Checklists) > 0 {
								fmt.Printf("[Trello Migration] Converted %d checklists from card %s\n", len(card.Checklists), card.ID)
							}

							// Labels
							for _, label := range card.Labels {
								color, exists := trelloColorMap[label.Color]
								if !exists {
									fmt.Printf("[Trello Migration] Color %s not mapped for trello card %s, falling back to transparent\n", label.Color, card.ID)
									color = trelloColorMap["transparent"]
								}

								task.Labels = append(task.Labels, &models.Label{
									Title:    label.Name,
									HexColor: color,
								})

								fmt.Printf("[Trello Migration] Converted label %s from card %s\n", label.ID, card.ID)

							}
							if len(card.Attachments) > 0 {
								fmt.Printf("[Trello Migration] Downloading %d card attachments from card %s\n", len(card.Attachments), card.ID)
							}

							for _, attachment := range card.Attachments {
								if attachment.IsUpload {
									fmt.Printf("[Trello Migration] Downloading card attachment %s\n", attachment.ID)

									buf, err := migration.DownloadFileWithHeaders(attachment.URL, map[string][]string{
										"Authorization": {`OAuth oauth_consumer_key="` + trelloApiKey + `", oauth_token="` + trelloApiToken + `"`},
									})
									if err != nil {
										return nil, err
									}

									vikunjaAttachment := &models.TaskAttachment{
										File: &models.File{
											Name:        attachment.Name,
											Mime:        attachment.MimeType,
											Size:        uint64(buf.Len()),
											FileContent: buf.Bytes(),
										},
									}

									if card.IDAttachmentCover != "" && card.IDAttachmentCover == attachment.ID {
										vikunjaAttachment.ID = 42
										task.CoverImageAttachmentID = 42
									}
									task.Attachments = append(task.Attachments, vikunjaAttachment)
									fmt.Printf("[Trello Migration] Downloaded card attachment %s\n", attachment.ID)
									continue
								}

								task.Description += `<p><a href="` + attachment.URL + `">` + attachment.Name + "</a></p>\n"
							}

							// When the cover image was set manually, we need to add it as an attachment
							if card.ManualCoverAttachment && len(card.Cover.Scaled) > 0 {

								cover := card.Cover.Scaled[len(card.Cover.Scaled)-1]

								buf, err := migration.DownloadFile(cover.URL)
								if err != nil {
									return nil, err
								}

								coverAttachment := &models.TaskAttachment{
									ID: 43,
									File: &models.File{
										Name:        cover.ID + ".jpg",
										Mime:        "image/jpg", // Seems to always return jpg
										Size:        uint64(buf.Len()),
										FileContent: buf.Bytes(),
									},
								}

								task.Attachments = append(task.Attachments, coverAttachment)
								task.CoverImageAttachmentID = coverAttachment.ID
							}

							for _, action := range card.Actions {
								if action.DidCommentCard() {
									if task.Comments == nil {
										task.Comments = []*models.TaskComment{}
									}

									comment := &models.TaskComment{
										Comment: action.Data.Text,
										Created: action.Date,
										Updated: action.Date,
										TaskID:  task.ID,
									}

									comment.Comment = "*" + action.MemberCreator.FullName + "*:\n\n" + comment.Comment

									comment.Comment, _ = convertMarkdownToHTML(comment.Comment)
									task.Comments = append(task.Comments, comment)
								}
							}

							tasks = append(tasks, task)

							// Hard limits to tasks size to maxTaskSize
							// Creates a bucket for each 200 sized tasks.
							if len(tasks) >= maxTaskSize {
								bucket := &models.Bucket{
									ProjectID:         projectFromData.ID,
									ProjectViewID:     view.ID,
									Title:             fmt.Sprintf("Archived Tasks %d", len(buckets)+1),
									TasksWithComments: tasks,
								}

								project.Buckets = append(project.Buckets, bucket)

								tasks = []*models.TaskWithComments{}
							}
						}

					}

					// If there's still left from the hard limit
					if len(tasks) > 0 {
						bucket := &models.Bucket{
							ProjectID:         projectFromData.ID,
							ProjectViewID:     view.ID,
							Title:             fmt.Sprintf("Archived Tasks %d", len(buckets)+1),
							TasksWithComments: tasks,
						}

						project.Buckets = append(project.Buckets, bucket)
					}
				}
			}
			fmt.Printf("[Trello Migration] Converted all cards to tasks for board %s\n", board.ID)

			hierarchy = append(hierarchy, project)
		}
	}

	return hierarchy, nil
}

func convertMarkdownToHTML(input string) (output string, err error) {
	var buf bytes.Buffer
	err = goldmark.Convert([]byte(input), &buf)
	if err != nil {
		return
	}
	//#nosec - we are not responsible to escape this as we don't know the context where it is used
	return buf.String(), nil
}
