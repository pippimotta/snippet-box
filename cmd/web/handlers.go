package main

import (
	"errors"
	"fmt"

	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/pippimotta/snippet-box/internal/models"
	"github.com/pippimotta/snippet-box/internal/validator"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// The snippet part
func (a *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := a.snippets.Latest()
	if err != nil {
		a.serverError(w, err)
		return
	}

	data := a.newTemplateData(r)
	data.Snippets = snippets

	a.render(w, http.StatusOK, "home.tmpl", data)

}

func (a *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		a.notFound(w)
		return
	}

	snippet, err := a.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			a.notFound(w)
		} else {
			a.serverError(w, err)
		}
		return
	}

	data := a.newTemplateData(r)
	data.Snippet = snippet

	a.render(w, http.StatusOK, "view.tmpl", data)
}

func (a *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = &snippetCreateForm{
		Expires: 365,
	}
	a.render(w, http.StatusOK, "create.tmpl", data)
}

func (a *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	var form snippetCreateForm
	err := a.decodePostForm(r, &form)

	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	//use Validator check the title value is not blank & is not more than 100 characters long,
	//if fails, add a message to the erros maps using the field name as key
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")

	//check the Content value isn't blank
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	//check the expires value matches one of the permitted Values (365, 1, 7)
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must be equal 1, 7 or 365")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	id, err := a.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		a.serverError(w, err)
		return
	}

	//add the infomation of successfully posting new snippet into session
	a.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

//The userLogin/out part

func (a *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = &userSignupForm{}
	a.render(w, http.StatusOK, "signup.tmpl", data)
}

func (a *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	//validate the form contents
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	form.CheckField(validator.IsEmail(form.Email), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	//create a new user record in the database
	err = a.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			a.serverError(w, err)
		}
		return
	}

	//if no error, show the flash message of success
	a.sessionManager.Put(r.Context(), "flash", "Your Sign Up was successful. Please log in :)")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (a *application) userLogin(w http.ResponseWriter, r *http.Request) {

	data := a.newTemplateData(r)
	data.Form = &userLoginForm{}
	a.render(w, http.StatusOK, "login.tmpl", data)
}

func (a *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := a.decodePostForm(r, &form)

	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.IsEmail(form.Email), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
	}

	//check if the credentials are valid. if not add a generic non-field error

	id, err := a.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			a.serverError(w, err)
		}
		return
	}

	//use the RenewToken method on the current session to change the session ID
	//when the authentication state changes(login -> logout)
	err = a.sessionManager.RenewToken(r.Context())
	if err != nil {
		a.serverError(w, err)
		return
	}
	//add the ID of the current user to session so that this user is registered as "login"
	a.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (a *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := a.sessionManager.RenewToken(r.Context())
	if err != nil {
		a.serverError(w, err)
		return
	}
	//remove the authenticatedUserID from the session data so that the user is Logged out
	a.sessionManager.Remove(r.Context(),"authenticatedUserID")
	a.sessionManager.Put(r.Context(),"flash","You've been logged out successfully!")

	http.Redirect(w, r ,"/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("OK"))
}