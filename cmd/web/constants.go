package main

type contextKey string
type sessionKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")

const authenticatedUserID = sessionKey("authenticatedUserID")
const redirectAfterLogin = sessionKey("redirectAfterLogin")
const flash = sessionKey("flash")
