package screenscraper

// Response is satisfied by all generated *XxxResponse types
type Response interface {
	StatusCode() int
}

// IsNotFound returns true if the response is 404 Not Found
func IsNotFound(r Response) bool {
	return r.StatusCode() == 404
}

// IsQuotaExceeded returns true if quota exceeded (430 or 431)
func IsQuotaExceeded(r Response) bool {
	code := r.StatusCode()
	return code == 430 || code == 431
}

// IsRateLimited returns true if rate limited (429)
func IsRateLimited(r Response) bool {
	return r.StatusCode() == 429
}

// IsServerBusy returns true if server is overloaded (401)
// This occurs when API is closed for non-members due to high CPU usage (>60%)
func IsServerBusy(r Response) bool {
	return r.StatusCode() == 401
}

// IsInvalidCredentials returns true if credentials are invalid (403)
func IsInvalidCredentials(r Response) bool {
	return r.StatusCode() == 403
}

// IsAPILocked returns true if API is completely closed (423)
func IsAPILocked(r Response) bool {
	return r.StatusCode() == 423
}

// IsBlacklisted returns true if software is blacklisted (426)
func IsBlacklisted(r Response) bool {
	return r.StatusCode() == 426
}

// IsSuccess returns true if the response is 2xx
func IsSuccess(r Response) bool {
	code := r.StatusCode()
	return code >= 200 && code < 300
}
