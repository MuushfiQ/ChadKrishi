package main

import "github.com/gin-gonic/gin"

// Photo updates are independent of garden setup, so new members can use them too.
func (a *app) updateProfilePhoto(c *gin.Context) {
	var request struct {
		PhotoPath *string `json:"photoPath"`
	}
	if !decodeJSON(c, &request) {
		return
	}
	if request.PhotoPath == nil {
		writeError(c, 400, "photoPath is required; use an empty string to remove your photo")
		return
	}
	if !a.validImage(c, *request.PhotoPath) {
		return
	}
	_, err := a.db.ExecContext(c, "UPDATE users SET profile_photo=NULLIF(?,'') WHERE id=?", *request.PhotoPath, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	user, err := a.loadProfile(c, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	c.JSON(200, gin.H{"user": user})
}
