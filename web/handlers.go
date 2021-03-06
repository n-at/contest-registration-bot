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

type notificationRequest struct {
	Id      uint64 `form:"notification_id"`
	Message string `form:"message"`
}

type idRequest struct {
	Id uint64 `form:"id" param:"id" query:"id"`
}

type participantIdRequest struct {
	ContestId     uint64 `form:"id" param:"id" query:"id"`
	ParticipantId uint64 `form:"participant_id" param:"participant_id" query:"participant_id"`
}

type notificationIdRequest struct {
	ContestId      uint64 `form:"id" param:"id" query:"id"`
	NotificationId uint64 `form:"notification_id" param:"notification_id" query:"notification_id"`
}

///////////////////////////////////////////////////////////////////////////////
//contests

// contestsGet List all contests
func contestsGet(c echo.Context) error {
	contests, err := storage.GetContests()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "templates/contests.twig", pongo2.Context{
		"contests": contests,
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
	if err := (&echo.DefaultBinder{}).BindBody(c, &contestData); err != nil {
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
	var err error

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

	if err := storage.SaveContest(contest); err != nil {
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

	if err := storage.SaveContest(contest); err != nil {
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

	if err := storage.SaveContest(contest); err != nil {
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

	if err := csvWriter.Write([]string{"login", "password", "name"}); err != nil {
		return err
	}

	for _, participant := range participants {
		if err := csvWriter.Write([]string{participant.Login, participant.Password, participant.Name}); err != nil {
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
	if err := (&echo.DefaultBinder{}).BindBody(c, &participantData); err != nil {
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

	if err := storage.SaveContestParticipant(participant); err != nil {
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

	if err := storage.DeleteContestParticipant(participant.Id); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contest/%d/participants", contestId))
}

func contestParticipant(c echo.Context) (*storage.ContestParticipant, error) {
	var id participantIdRequest
	if err := (&echo.DefaultBinder{}).Bind(&id, c); err != nil {
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

// contestNotifications List contest notifications
func contestNotifications(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	notifications, err := storage.GetContestNotifications(contest.Id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "templates/notifications.twig", pongo2.Context{
		"contest":       contest,
		"notifications": notifications,
	})
}

// contestNotificationNew Form for new contest notification
func contestNotificationNew(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "templates/notification.twig", pongo2.Context{
		"contest":      contest,
		"notification": nil,
	})
}

// contestNotificationEdit Edit existing contest notification
func contestNotificationEdit(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	notification, err := contestNotification(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "templates/notification.twig", pongo2.Context{
		"contest":      contest,
		"notification": notification,
	})
}

// contestNotificationSave Save contest notification
func contestNotificationSave(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	var notificationData notificationRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &notificationData); err != nil {
		return err
	}
	if len(notificationData.Message) == 0 {
		return errors.New("notification message required")
	}

	var notification *storage.ContestNotification

	if notificationData.Id != 0 {
		notification, err = storage.GetContestNotification(notificationData.Id)
		if err != nil {
			return err
		}
		if notification.ContestId != contest.Id {
			return errors.New("notification belongs to other contest")
		}
		notification.Message = notificationData.Message
	} else {
		notification = &storage.ContestNotification{
			ContestId: contest.Id,
			Message:   notificationData.Message,
		}
	}

	if err := storage.SaveContestNotification(notification); err != nil {
		return err
	}
	if err := registrationBot.SendNotifications(contest.Id, notification.Message); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contest/%d/notifications", contest.Id))
}

// contestNotificationDelete Delete given contest notification
func contestNotificationDelete(c echo.Context) error {
	contest, err := contest(c)
	if err != nil {
		return err
	}

	notification, err := contestNotification(c)
	if err != nil {
		return err
	}

	if err := storage.DeleteContestNotification(notification.Id); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contest/%d/notifications", contest.Id))
}

func contestNotification(c echo.Context) (*storage.ContestNotification, error) {
	var notificationId notificationIdRequest
	if err := (&echo.DefaultBinder{}).Bind(&notificationId, c); err != nil {
		return nil, err
	}

	notification, err := storage.GetContestNotification(notificationId.NotificationId)
	if err != nil {
		return nil, err
	}
	if notification.ContestId != notificationId.ContestId {
		return nil, errors.New("notification belongs to other contest")
	}

	return notification, nil
}
