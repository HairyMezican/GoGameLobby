package controllers

import (
	"../controller"
	"../login"
	"../models"
	"../models/user"
	"time"
)

type UserController struct {
	controller.Heart
}

func (this UserController) Index() {
	var users []user.User
	err := user.U.AllUsers(&users)
	if err != nil {
		panic(err)
	}

	this.Set("Users", users)
	this.Set("Title", "Users")
}

func (this UserController) Show() {
	u := this.GetVar("User").(*user.User)

	this.Set("Title", u.ClashTag)
}

func (this UserController) New() {
	authorization, isString := this.Session().Clear("authorization").(string)
	if !isString {
		this.NotAuthorized()
		return
	}

	this.Set("authorization", authorization)
	this.Set("access", this.Session().Clear("access"))
	this.Set("refresh", this.Session().Clear("refresh"))
	this.Set("expiry", this.Session().Clear("expiry"))
	this.Set("auth_id", this.Session().Clear("auth_id"))

	this.Set("Title", "New User")
}

func (this UserController) Update() {
	u := this.GetVar("User").(*user.User)
	current := this.GetVar("CurrentUser").(*user.User)
	switch this.GetFormValue("Action") {
	case "FriendRequest":
		current.RequestFriend(u.ClashTag)
	case "DenyRequest":
		current.DenyRequest(u.ClashTag)
	case "Unfriend":
		current.Unfriend(u.ClashTag)
	}
}

func (this UserController) Create() {
	var authData user.AuthorizationData
	var err error

	//would be nice to replace this with some kind of reflection based reader
	authData.Authorization = this.GetFormValue("User[Authorization][Type]")
	authData.Id = this.GetFormValue("User[Authorization][ID]")
	authData.Token.AccessToken = this.GetFormValue("User[Authorization][Access]")
	authData.Token.RefreshToken = this.GetFormValue("User[Authorization][Refresh]")
	authData.Token.Expiry, err = time.Parse(time.RFC1123, this.GetFormValue("User[Authorization][Expiry]"))
	if err != nil {
		panic("Can't Convert Expiry to Time")
	}

	defer func() {
		rec := recover()
		if rec != nil {
			this.AddFlash("I'm sorry, but there was an error with your form")
			this.FinishWithMiddleware(login.NewUserForm(authData))
		}
	}()

	u := user.NewUser()

	u.ClashTag = urlify(this.GetFormValue("User[ClashTag]"))
	u.Authorizations = []user.AuthorizationData{authData}

	errs := model.Save(u)
	if errs != nil {
		panic(errs)
	}

	(login.V)(this.Vars).LogIn(u)
	this.RespondWith(u)
}

func init() {
	controller.RegisterController(&UserController{}, "users", "User", func(s string, vars map[string]interface{}) (interface{}, bool) {
		result := user.U.UserFromClashTag(s)
		return result, result != nil
	}).AddTo(Root)
}
