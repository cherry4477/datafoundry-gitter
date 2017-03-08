# datafoundry-gitter


Then in the AuthHandler I saved the token details before requesting the client

tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
if err != nil {
    c.AbortWithError(http.StatusBadRequest, err)
    return
}
// save the token
session.Set("AccessToken", tok.AccessToken)
session.Set("RefreshToken", tok.RefreshToken)
session.Set("TokenType", tok.TokenType)
session.Set("Expiry", tok.Expiry.Format(time.RFC3339))
session.Save()


client := conf.Client(oauth2.NoContext, tok)
Now I can use that info in other handlers to recreate the client and make API calls.

func ListStuff(c *gin.Context) {


	ctx := context.Background()
	session := sessions.Default(c)


	token := new(oauth2.Token)
	token.AccessToken = session.Get("AccessToken").(string)
	token.RefreshToken = session.Get("RefreshToken").(string)
	token.RefreshToken = session.Get("RefreshToken").(string)
	t := session.Get("Expiry").(string)
	token.Expiry, _ = time.Parse(time.RFC3339, t)
	token.TokenType = session.Get("TokenType").(string)


	client := conf.Client(ctx, token)

    // ...
}