package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/warrenwingaru/go-trello"
	"os"
)

var apiKey string
var apiToken string

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
	apiKey = os.Getenv("TRELLO_API_KEY")
	apiToken = os.Getenv("TRELLO_API_TOKEN")
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	trelloClient := trello.NewClient(apiKey, apiToken)
	trelloClient.Logger = logger

	boards, err := getTrelloBoards(trelloClient)
	if err != nil {
		panic(err)
	}

	organizationMap := getTrelloOrganizationsWithBoards(boards)
	for organizationID, boards := range organizationMap {
		logrus.Debugf("[Trello Migration] Getting organization with id %s\n", organizationID)
		orgName := organizationID
		if orgName != "Personal" {
			organization, err := trelloClient.GetOrganization(organizationID, trello.Defaults())
			if err != nil {
				panic(err)
			}
			orgName = organization.DisplayName
		}

		for _, board := range boards {
			logrus.Debugf("[Trello Migration] Getting card data for board %s for organization %s\n", board.ID, organizationID)

			err = fillCardData(trelloClient, board)
			if err != nil {
				panic(err)
			}
			logrus.Debugf("[Trello Migration] Got card data for board %s for organization %s\n", board.ID, organizationID)
		}

		logrus.Debugf("[Trello Migration] Start conreting trello data for organization %s\n", organizationID)

		//hiararchy, err := convertTrelloToVikunja(boards, vikunjaData)
		//if err != nil {
		//	panic(err)
		//}
		//hiarachies = append(hiarachies, hiararchy)
	}

	jsonData, err := json.Marshal(boards)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("trello.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

}

func getTrelloBoards(client *trello.Client) (trelloData []*trello.Board, err error) {
	logrus.Println("[Trello Migration] Getting boards...")

	trelloData, err = client.GetMyBoards(trello.Defaults())
	if err != nil {
		return nil, err
	}

	logrus.Debugf("[Trello Migration] Got %d trello boards\n", len(trelloData))

	return
}

func getTrelloOrganizationsWithBoards(boards []*trello.Board) (boardsByOrg map[string][]*trello.Board) {

	boardsByOrg = make(map[string][]*trello.Board)

	for _, board := range boards {
		// Trello boards without an organization are considered personal boards
		if board.IDOrganization == "" {
			board.IDOrganization = "Personal"
		}

		_, has := boardsByOrg[board.IDOrganization]
		if !has {
			boardsByOrg[board.IDOrganization] = []*trello.Board{}
		}

		boardsByOrg[board.IDOrganization] = append(boardsByOrg[board.IDOrganization], board)
	}

	return
}

func fillCardData(client *trello.Client, board *trello.Board) (err error) {
	allArg := trello.Arguments{"fields": "all"}

	logrus.Debugf("[Trello Migration] Getting projects for board %s\n", board.ID)

	board.Lists, err = board.GetLists(trello.Defaults())
	if err != nil {
		return err
	}

	logrus.Debugf("[Trello Migration] Got %d projects for board %s\n", len(board.Lists), board.ID)

	listMap := make(map[string]*trello.List, len(board.Lists))
	for _, list := range board.Lists {
		listMap[list.ID] = list
	}

	logrus.Debugf("[Trello Migration] Getting cards for board %s\n", board.ID)

	cards, err := board.GetFilteredCards("closed", allArg)
	if err != nil {
		return
	}

	logrus.Debugf("[Trello Migration] Got %d cards for board %s\n", len(cards), board.ID)

	for _, card := range cards {
		list, exists := listMap[card.IDList]
		if !exists {
			continue
		}

		if card.Badges.Attachments > 0 {
			card.Attachments, err = card.GetAttachments(allArg)
			if err != nil {
				return
			}
		}

		if card.Badges.Comments > 0 {
			card.Actions, err = card.GetCommentActions()
			if err != nil {
				return
			}
		}

		if len(card.IDCheckLists) > 0 {
			for _, checkListID := range card.IDCheckLists {
				checklist, err := client.GetChecklist(checkListID, allArg)
				if err != nil {
					return err
				}

				checklist.CheckItems = []trello.CheckItem{}
				err = client.Get("checklists/"+checkListID+"/checkItems", allArg, &checklist.CheckItems)
				if err != nil {
					return err
				}

				card.Checklists = append(card.Checklists, checklist)
				logrus.Debugf("Retrieved checklist %s for card %s\n", checkListID, card.ID)
			}
		}

		list.Cards = append(list.Cards, card)
	}

	logrus.Debugf("[Trello Migration] Looked for attachements on all cards of board %s\n", board.ID)

	return
}
