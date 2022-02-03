package web

import (
	"contest-registration-bot/storage"
	"errors"
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
//contests

type contestRequest struct {
	Id          uint64 `form:"id"`
	Name        string `form:"name"`
	Description string `form:"description"`
	When        string `form:"when"`
	Where       string `form:"where"`
}

type idRequest struct {
	Id uint64 `form:"id" param:"id" query:"id"`
}

// contestsGet List all contests
func contestsGet(c echo.Context) error {
	contests, err := storage.GetContests()
	return c.Render(http.StatusOK, "templates/contests_index.twig", pongo2.Context{
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
	var id idRequest
	err := (&echo.DefaultBinder{}).Bind(&id, c)
	if err != nil {
		return err
	}

	contest, err := storage.GetContest(id.Id)
	if err != nil {
		return err
	}
	if contest == nil {
		return errors.New("contest not found")
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

func contestUpdateClosed(value bool, c echo.Context) error {
	var id idRequest
	err := (&echo.DefaultBinder{}).Bind(&id, c)
	if err != nil {
		return err
	}

	contest, err := storage.GetContest(id.Id)
	if err != nil {
		return err
	}
	if contest == nil {
		return errors.New("contest not found")
	}

	contest.Closed = value

	err = storage.SaveContest(contest)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

// contestHide Hide contest from participants
func contestHide(c echo.Context) error {
	return contestUpdateHidden(true, c)
}

// contestShow Show contest for participants
func contestShow(c echo.Context) error {
	return contestUpdateHidden(false, c)
}

func contestUpdateHidden(value bool, c echo.Context) error {
	var id idRequest
	err := (&echo.DefaultBinder{}).Bind(&id, c)
	if err != nil {
		return err
	}

	contest, err := storage.GetContest(id.Id)
	if err != nil {
		return err
	}
	if contest == nil {
		return errors.New("contest not found")
	}

	contest.Hidden = value

	err = storage.SaveContest(contest)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

///////////////////////////////////////////////////////////////////////////////
//contest participants

//TODO

///////////////////////////////////////////////////////////////////////////////
//contest notifications

//TODO
