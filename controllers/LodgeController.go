package controllers

import (
	"../controller"
	"../login"
	"../models"
	"../models/game"
	"../models/lodge"
)

type LodgeController struct {
	controller.Heart
}

func (this LodgeController) Index() {
	var lodges []lodge.Lodge
	err := lodge.L.AllLodges(&lodges)
	if err != nil {
		panic(err)
	}

	this.Set("Lodges", lodges)
	this.Set("Title", "Mason Lodges")
}

func (this LodgeController) Show() {
	l, isLodge := this.GetVar("Lodge").(*lodge.Lodge)
	if !isLodge {
		panic("Can't find lodge")
	}

	inProgress := game.G.UnpublishedGamesFromLodge(l)
	if len(inProgress) > 0 {
		this.Set("InProgress", inProgress)
	}

	this.Set("Title", l.Name)
}

func (this LodgeController) New() {
	this.Set("Title", "Create a Mason Lodge")
}

func (this LodgeController) Create() {
	defer func() {
		rec := recover()
		if rec != nil {
			errs, isSlice := rec.([]error)
			if isSlice {
				for _, err := range errs {
					this.AddFlash(err.Error())
				}
			}
			err, isError := rec.(error)
			if isError {
				this.AddFlash(err.Error())
			}
			str, isStr := rec.(string)
			if isStr {
				this.AddFlash(str)
			}
			this.AddFlash("You fucked something up, please try again")
			this.RedirectTo("/lodges/new")
		}
	}()

	l := lodge.NewLodge()

	l.Name = urlify(this.GetFormValue("Lodge[Name]"))
	user, loggedIn := (login.V)(this.Vars).CurrentUser()
	if !loggedIn {
		panic("you must be logged in to create a mason lodge")
	}
	l.AddMason(user)

	errs := model.Save(l)
	if errs != nil {
		panic(errs)
	}

	this.RespondWith(l)
}

var Lodge *controller.ControllerShell

func init() {
	Lodge = controller.RegisterController(new(LodgeController), "lodges", "Lodge", func(s string, vars map[string]interface{}) (interface{}, bool) {
		result := lodge.L.LodgeFromName(s)
		return result, result != nil
	})
	Lodge.AddTo(Root)
}
