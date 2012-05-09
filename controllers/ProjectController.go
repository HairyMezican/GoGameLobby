package controllers

import (
	"../controller"
	"../login"
	"../models"
	"../models/game"
	"../models/lodge"
)

func init() {
	controller.RegisterController(&ProjectController{}).AddAsSubresource(Lodge)
}

type ProjectController struct {
	controller.Heart
}

func (this ProjectController) Indexer(query string) (interface{}, bool) {
	l, isLodge := this.Get("Lodge").(*lodge.Lodge)
	if !isLodge {
		panic("Cannot find lodge")
	}

	result := game.G.GameFromLodgeAndName(l, query)
	return result, result != nil
}

func (ProjectController) RouteName() string {
	return "projects"
}

func (ProjectController) VarName() string {
	return "Game"
}

func (this ProjectController) Show() controller.Response {
	g, isGame := this.Get("Game").(*game.Game)
	if !isGame {
		panic("Can't find Game")
	}
	currentUser, loggedIn := login.CurrentUser(this.Vars)
	if g.Live && loggedIn {
		gameModes := g.GetGameModes(currentUser)
		this.Set("GameModes", gameModes)
	}

	this.Set("Title", g.Name)
	return this.DefaultResponse()
}

func (this ProjectController) Create() (response controller.Response) {
	l, isLodge := this.Get("Lodge").(*lodge.Lodge)
	if !isLodge {
		panic("Cannot find lodge")
	}

	defer func() {
		rec := recover()
		if rec != nil {
			response = this.Redirection(l.Url())
			this.AddFlash("You fucked something up, please try again")

		}
	}()

	g := game.NewGame()
	g.Name = urlify(this.GetFormValue("Game[Name]"))
	g.Lodge = l.Name

	errs := model.Save(g)
	if errs != nil {
		panic(errs)
	}

	return this.RespondWith(g)
}
