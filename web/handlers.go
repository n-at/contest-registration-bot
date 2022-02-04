package web

import (
	"contest-registration-bot/storage"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type contestRequest struct {
	Id          uint64 `form:"id"`
	Name        string `form:"name"`
	Description string `form:"description"`
	When        string `form:"when"`
	Where       string `form:"where"`
}

type participantRequest struct {
	Id        uint64 `form:"participant_id"`
	Name      string `form:"name"`
	School    string `form:"school"`
	Contacts  string `form:"contacts"`
	Languages string `form:"languages"`
	Login     string `form:"login"`
	Password  string `form:"password"`
}

type idRequest struct {
	Id uint64 `form:"id" param:"id" query:"id"`
}

type participantIdRequest struct {
	ContestId     uint64 `form:"id" param:"id" query:"id"`
	ParticipantId uint64 `form:"participant_id" param:"participant_id" query:"participant_id"`
}

///////////////////////////////////////////////////////////////////////////////
//contests

// contestsGet List all contests
func contestsGet(c echo.Context) error {
	contests, err := storage.GetContests()
	return c.Render(http.StatusOK, "templates/contests.twig", pongo2.Context{
		"contests":   contests,
		"page_error": err,
	})
}

// contestNew Form to create new contest
func contestNew(c echo.Context) error {
	return c.Render(http.StatusOK, "templates/contest.twig", pongo2.Context{
		"contest": nil,
	})
}

// contestGet Form to edit existing contest
func contestGet(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "templates/contest.twig", pongo2.Context{
		"contest": contest,
	})
}

// contestSave Save new or update existing contest
func contestSave(c echo.Context) error {
	var contestData contestRequest
	err := (&echo.DefaultBinder{}).BindBody(c, &contestData)
	if err != nil {
		return err
	}
	if len(contestData.Name) == 0 {
		return errors.New("contest name required")
	}
	if len(contestData.Description) == 0 {
		return errors.New("contest description required")
	}
	if len(contestData.Where) == 0 {
		return errors.New("contest location required")
	}
	if len(contestData.When) == 0 {
		return errors.New("contest date required")
	}

	var contest *storage.Contest

	if contestData.Id != 0 {
		contest, err = storage.GetContest(contestData.Id)
		if err != nil {
			return err
		}
		contest.Name = contestData.Name
		contest.Description = contestData.Description
		contest.When = contestData.When
		contest.Where = contestData.Where
	} else {
		contest = &storage.Contest{
			Name:        contestData.Name,
			Description: contestData.Description,
			When:        contestData.When,
			Where:       contestData.Where,
			Closed:      false,
			Hidden:      false,
		}
	}

	err = storage.SaveContest(contest)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

// contestClose Close contest for registration
func contestClose(c echo.Context) error {
	return contestUpdateClosed(true, c)
}

// contestOpen Open contest for participation
func contestOpen(c echo.Context) error {
	return contestUpdateClosed(false, c)
}

// contestHide Hide contest from participants
func contestHide(c echo.Context) error {
	return contestUpdateHidden(true, c)
}

// contestShow Show contest for participants
func contestShow(c echo.Context) error {
	return contestUpdateHidden(false, c)
}

func contestUpdateClosed(value bool, c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	contest.Closed = value

	err = storage.SaveContest(contest)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func contestUpdateHidden(value bool, c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	contest.Hidden = value

	err = storage.SaveContest(contest)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func contest(c echo.Context) (*storage.Contest, error) {
	var id idRequest
	err := (&echo.DefaultBinder{}).Bind(&id, c)
	if err != nil {
		return nil, err
	}

	contest, err := storage.GetContest(id.Id)
	if err != nil {
		return nil, err
	}
	if contest == nil {
		return nil, errors.New("contest not found")
	}
	return contest, nil
}

///////////////////////////////////////////////////////////////////////////////
//contest participants

// participantsList List all contest participants
func participantsList(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	participants, err := storage.GetContestParticipants(contest.Id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "templates/participants.twig", pongo2.Context{
		"contest":      contest,
		"participants": participants,
	})
}

// participantsExport Export contest participant to CSV
func participantsExport(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	participants, err := storage.GetContestParticipants(contest.Id)
	if err != nil {
		return err
	}

	stringBuilder := &strings.Builder{}
	csvWriter := csv.NewWriter(stringBuilder)
	csvWriter.Comma = ';'
	csvWriter.UseCRLF = false

	err = csvWriter.Write([]string{"login", "password", "name"})
	if err != nil {
		return err
	}

	for _, participant := range participants {
		err = csvWriter.Write([]string{participant.Login, participant.Password, participant.Name})
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()

	c.Response().Header().Set("Content-Disposition", "attachment; filename=\"participants.csv\"")
	return c.Blob(http.StatusOK, "text/csv", []byte(stringBuilder.String()))
}

// participantNew Form to create new participant
func participantNew(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "templates/participant.twig", pongo2.Context{
		"contest":     contest,
		"participant": nil,
	})
}

// participantEdit Form to edit existing participant
func participantEdit(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}
	participant, err := contestParticipant(c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "templates/participant.twig", pongo2.Context{
		"contest":     contest,
		"participant": participant,
	})
}

// participantSave Save new or update existing participant
func participantSave(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	var participantData participantRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &participantData)
	if err != nil {
		return err
	}
	if len(participantData.Name) == 0 {
		return errors.New("participant name required")
	}

	var participant *storage.ContestParticipant

	if participantData.Id != 0 {
		participant, err = storage.GetContestParticipant(participantData.Id)
		if err != nil {
			return err
		}
		if participant.ContestId != contest.Id {
			return errors.New("participant does not belong to contest")
		}
		participant.Name = participantData.Name
		participant.School = participantData.School
		participant.Contacts = participantData.Contacts
		participant.Languages = participantData.Languages
		participant.Login = participantData.Login
		participant.Password = participantData.Password
	} else {
		participant = &storage.ContestParticipant{
			ContestId: contest.Id,
			Name:      participantData.Name,
			School:    participantData.School,
			Contacts:  participantData.Contacts,
			Languages: participantData.Languages,
			Login:     participantData.Login,
			Password:  participantData.Password,
		}
	}

	err = storage.SaveContestParticipant(participant)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contest/%d/participants", contest.Id))
}

// participantDelete Delete participant
func participantDelete(c echo.Context) error {
	participant, err := contestParticipant(c)
	if err != nil {
		return err
	}

	contestId := participant.ContestId

	err = storage.DeleteContestParticipant(participant.Id)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contest/%d/participants", contestId))
}

func contestParticipant(c echo.Context) (*storage.ContestParticipant, error) {
	var id participantIdRequest
	err := (&echo.DefaultBinder{}).Bind(&id, c)
	if err != nil {
		return nil, err
	}

	participant, err := storage.GetContestParticipant(id.ParticipantId)
	if err != nil {
		return nil, err
	}
	if participant.ContestId != id.ContestId {
		return nil, errors.New("participant does not belong to contest")
	}

	return participant, nil
}

///////////////////////////////////////////////////////////////////////////////
//contest notifications

//TODO
