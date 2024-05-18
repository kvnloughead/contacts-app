package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/kvnloughead/contacts-app/internal/models"
	"github.com/kvnloughead/contacts-app/internal/validator"
)

//
// Basic Handlers (ping, home, about)
//

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Displays home page in response to GET /. If we were using http.ServeMux we
// would have to check the URL, but with httprouter.Router, "/" is exclusive.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	contacts, err := app.contacts.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Contacts = contacts

	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

// Displays about page in response to GET /about.
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "about.tmpl", data)
}

//
// Contact handlers
//

// contactFormFields struct contains the form fields for the
// /contacts/create or /contacts/edit forms.
type contactFormFields struct {
	ID                  int        `form:"id"`
	First               string     `form:"first"`
	Last                string     `form:"last"`
	Phone               string     `form:"phone"`
	Email               string     `form:"email"`
	Version             int        `form:"version"`
	validator.Validator `form:"-"` // "-" tells formDecoder to ignore the field
}

// View page for the contact with the given ID.
// If there's no matching contact a 404 NotFound response is sent.
func (app *application) contactView(w http.ResponseWriter, r *http.Request) {
	// Params are stored by httprouter in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// Once parsed, params are available by params.ByName().
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	contact, err := app.contacts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Contact = contact

	app.render(w, r, http.StatusOK, "view.tmpl", data)
}

func (app *application) contactCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = contactFormFields{}
	app.render(w, r, http.StatusOK, "create.tmpl", data)
}

/*
Inserts a new record into the database. If successful, redirects the user to
the corresponding page with a 303 status code.

If one or more fields are invalid, the form is rendered again with a 422 status
code, displaying the appropriate error messages.

If we were using http.ServeMux, we would have to check the method in this handler.
*/
func (app *application) contactCreatePost(w http.ResponseWriter, r *http.Request) {
	// Create an instance of our form struct and decode it with decodePostForm.
	// This automatically parses the values passed as the second argument into the
	// corresponding struct fields, making appropriate data conversions.
	var form contactFormFields
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate all form fields.
	form.CheckField(validator.NotBlank(form.First), "first", "This field can't be blank.")
	form.CheckField(validator.MaxChars(form.First, 100), "first", "This can't contain more than 100 characters.")
	form.CheckField(validator.NotBlank(form.Last), "last", "This field can't be blank.")
	form.CheckField(validator.MaxChars(form.Last, 100), "last", "This can't contain more than 100 characters.")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field can't be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email.")

	form.CheckField(validator.NotBlank(form.Phone), "phone", "This field can't be blank.")
	form.CheckField(validator.ValidatePhoneNumberInput(form.Phone), "phone", "Invalid phone number.")

	// If there are any validation errors, render the page again with the errors.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// Insert new record or respond with a server error.
	id, err := app.contacts.Insert(form.First, form.Last, form.Phone, form.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Assign text to session data with the key "flash". The data is stored in the
	// request's context. If there is no current session, a new one will be created.
	// The flash is added to our template data via the newTemplateData function.
	app.sessionManager.Put(r.Context(), string(flash), "Contact successfully created!")

	// Redirect to page containing the new contact.
	http.Redirect(w, r, fmt.Sprintf("/contacts/view/%d", id), http.StatusSeeOther)
}

// contactEdit handles GET /contacts/edit/:id requests by displaying a form to
// edit the contact.
func (app *application) contactEdit(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	contact, err := app.contacts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Contact = contact
	data.Form = contactFormFields{}

	app.render(w, r, http.StatusOK, "edit.tmpl", data)
}

// Updates a contact record. If successful, redirects the user to the
// corresponding view page with a 303 status code.
//
// If one or more fields are invalid, the form is rendered again with a 422
// status code, displaying the appropriate error messages.
//
// Uses POST because HTML forms don't support PUT.
func (app *application) contactEditPost(w http.ResponseWriter, r *http.Request) {
	var form contactFormFields
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate all form fields.
	form.CheckField(validator.NotBlank(form.First), "first", "This field can't be blank.")
	form.CheckField(validator.MaxChars(form.First, 100), "first", "This can't contain more than 100 characters.")
	form.CheckField(validator.NotBlank(form.Last), "last", "This field can't be blank.")
	form.CheckField(validator.MaxChars(form.Last, 100), "last", "This can't contain more than 100 characters.")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field can't be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email.")

	form.CheckField(validator.NotBlank(form.Phone), "phone", "This field can't be blank.")
	form.CheckField(validator.ValidatePhoneNumberInput(form.Phone), "phone", "Invalid phone number.")

	// If there are any validation errors, render the page again with the errors.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "edit.tmpl", data)
		return
	}

	var contact = models.Contact{ID: form.ID, First: form.First, Last: form.Last, Phone: form.Phone, Email: form.Email, Version: int32(form.Version)}

	// Update record or respond with a server error.
	err = app.contacts.Update(&contact)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.sessionManager.Put(r.Context(), "flash", "Another user has updated this contact. Please reload and try again.")
			http.Redirect(w, r, fmt.Sprintf("/contacts/edit/%d", form.ID), http.StatusSeeOther)
			return
		default:
			app.serverError(w, r, err)
			return
		}
	}

	// Assign text to session data with the key "flash". The data is stored in the
	// request's context. If there is no current session, a new one will be created.
	// The flash is added to our template data via the newTemplateData function.
	app.sessionManager.Put(r.Context(), string(flash), "Contact successfully updated!")

	// Redirect to page containing the new contact.
	http.Redirect(w, r, fmt.Sprintf("/contacts/view/%d", form.ID), http.StatusSeeOther)
}

func (app *application) contactDelete(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	// Once parsed, params are available by params.ByName().
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Get the contact from the database, if it exists.
	contact, err := app.contacts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Create template data and add the contact to it.
	data := app.newTemplateData(r)
	data.Contact = contact
	data.DeleteForm = true

	app.render(w, r, http.StatusOK, "view.tmpl", data)
}
