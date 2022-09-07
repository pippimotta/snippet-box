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
	Title   string `form:"title"`
	Content string	`form:"content"`
	Expires int	`form:"expires"`
	validator.Validator `form:"-"`
}

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
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must be equal 1, 7 or 365")

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

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
