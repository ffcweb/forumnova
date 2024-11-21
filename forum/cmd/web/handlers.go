package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"net/http"
	"strconv"

	"forum/internal/validator"
)

type accountCreateForm struct {
	Username string
	Email    string
	Password string
	validator.Validator
}

type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

type threadCreateForm struct {
	Title    string
	AuthorID int
	validator.Validator
}

type messageCreateForm struct {
	Body     string
	AuthorID int
	ThreadID int
	validator.Validator
}

// home shows the 10 latest threads.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	threads, err := app.threads.Latests()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Threads = threads
	app.render(w, r, http.StatusOK, "home", data)
}

// accountCreate shows a form the create an account.
func (app *application) accountCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountCreateForm{}
	app.render(w, r, http.StatusOK, "account-create", data)
}

// accountCreatePOST creates an account with the info in the POST request.
func (app *application) accountCreatePOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form := accountCreateForm{
		Username: r.PostForm.Get("username"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Username), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.ValidateEmail(form.Email), "email", "This field must be an email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.CheckPassword(form.Password), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "account-create", data)
		return
	}

	isEailExists, _ := app.users.Exists(form.Email)
	if isEailExists {
		data := app.newTemplateData(r)
		form.AddFieldError("email", "Sorry, this email is aready in use, Plese try another. ")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "account-create", data)
		return
	}

	id, err := app.users.Insert(form.Username, form.Email, form.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "flash", " Your signup was successful.")
	http.Redirect(w, r, fmt.Sprintf("/account/view/%d", id), http.StatusSeeOther)
}

// accountView shows information about an account.
func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	flash := app.sessionManager.PopString(r.Context(), "flash")
	data := app.newTemplateData(r)
	data.User = user
	data.Flash = flash
	app.render(w, r, http.StatusOK, "account-view", data)
}

// threadCreate shows a form to create a thread.
func (app *application) threadCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = threadCreateForm{}
	app.render(w, r, http.StatusOK, "thread-create", data)
}

// threadCreatePOST creates a post with the info in the POST request.
func (app *application) threadCreatePOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form := threadCreateForm{
		Title: r.PostForm.Get("title"),
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "thread-create", data)
		return
	}

	authorID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	id, err := app.threads.Insert(form.Title, authorID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Thread successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/thread/view/%d", id), http.StatusSeeOther)
}

// threadView shows a thread.
func (app *application) threadView(w http.ResponseWriter, r *http.Request) {
	idSegment := r.PathValue("id")
	id, err := strconv.Atoi(idSegment)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	thread, err := app.threads.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Thread = thread
	fmt.Printf("%+v\n", data.Thread)
	app.render(w, r, http.StatusOK, "thread-view", data)
}

// postCreate shows a form to create a message.
func (app *application) postCreate(w http.ResponseWriter, r *http.Request) {
	idSegment := r.PathValue("id")
	id, err := strconv.Atoi(idSegment)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	data := app.newTemplateData(r)
	data.ThreadID = id
	data.Form = messageCreateForm{}
	app.render(w, r, http.StatusOK, "message-create", data)
}

// postCreatePOST creates a message with the info in the POST request.
func (app *application) postCreatePOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	idSegment := r.PathValue("id")
	threadId, err := strconv.Atoi(idSegment)
	if err != nil || threadId < 1 {
		http.NotFound(w, r)
		return
	}

	form := messageCreateForm{
		Body: r.PostForm.Get("body"),
	}

	form.CheckField(validator.NotBlank(form.Body), "body", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Body, 1000), "body", "This field cannot be more than 1000 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.ThreadID = threadId
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "message-create", data)
		return
	}

	authorID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	postID, err := app.posts.Insert(form.Body, threadId, authorID)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Message %d successfully created!", postID))
	http.Redirect(w, r, fmt.Sprintf("/thread/view/%d", threadId), http.StatusSeeOther)
}

// userLogin initializes and displays the login form.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login", data)
}

// userLoginPost processes the login form, validates credentials, handles errors,
// and redirects on success.
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form := userLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("generic", "Email or password incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, fmt.Sprintf("/account/view/%d", id), http.StatusSeeOther)
}

// userLogoutPost handles user logout, renews the session token, and sets a logout
// message before redirecting.
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
