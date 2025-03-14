# Gin

## example

```
r := e2gin.DefaultEngine(&e2gin.Option{
DisableGzip: false,
StaticFiles: []*e2gin.StaticFiles{
{
FS:       e2exec.Must(fs.Sub(webFS, "mycash-web/build")),
HttpPath: "/",
},
{
FS:       assets.EmbedAssets,
HttpPath: "/assets",
},
},
HTMLTemplate: e2exec.Must(e2gin.ParseTemplates(templates.EmbedTemplates, tf)),
})

r.Use(cors.New(cors.Config{
AllowAllOrigins:           true,
AllowMethods:              []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
AllowHeaders:              []string{"Origin", "Content-Type", "Authorization", "Range", "X-Api-Consumer"},
ExposeHeaders:             []string{"Content-Range", "X-Total-Count"},
AllowCredentials:          true,
MaxAge:                    12 * time.Hour,
AllowWebSockets:           true,
AllowFiles:                true,
OptionsResponseStatusCode: 200,
}))

apiGroup := r.Group("/api/v1")
common.New(app.Instance).Routers(apiGroup)
```
