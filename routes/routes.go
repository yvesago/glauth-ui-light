package routes

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/md5" //nolint:gosec // only for cache headers

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"

	"github.com/gin-contrib/static"
	csrf "github.com/utrack/gin-csrf"

	"github.com/ulule/limiter"
	mgin "github.com/ulule/limiter/drivers/middleware/gin"
	"github.com/ulule/limiter/drivers/store/memory"

	"glauth-ui-light/config"
	. "glauth-ui-light/handlers"
	. "glauth-ui-light/helpers"
)

//go:embed web/assets/*
var server embed.FS

//go:embed web/templates/*
var templateFs embed.FS

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

func setConfig(cfg *config.WebConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Cfg", *cfg)
		c.Next()
	}
}

func LoadTls(port string) gin.HandlerFunc {
	return secure.New(secure.Config{
		SSLRedirect:           true,
		SSLHost:               port,
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline'; img-src 'self' data:",
		IENoOpen:              true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
	})
}

func initServer(cfg *config.WebConfig) *gin.Engine {
	r := gin.New()

	rate, e := limiter.NewRateFromFormatted("60-M") // 60 reqs/minute
	if e != nil {
		panic(e)
	}
	lStore := memory.NewStore()
	limitMiddleware := mgin.NewMiddleware(limiter.New(lStore, rate))

	// Set log config
	if !cfg.Debug {
		r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			// custom format
			// return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d \"%s\" %s\n",
			return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %q %s\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC3339),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		}))
	} else {
		r.Use(gin.Logger())
	}

	r.Use(gin.Recovery())

	// Set TrustedProxies
	r.SetTrustedProxies(cfg.Sec.TrustedProxies) //nolint:errcheck //useless check

	// Find source IP from proxy
	r.ForwardedByClientIP = true

	// Limit rate request
	r.Use(limitMiddleware)

	// SSL
	useSSL := false
	if cfg.SSL.Crt != "" {
		useSSL = true
	}
	r.Use(MiddlewareSession(useSSL))
	if useSSL {
		r.Use(LoadTls(cfg.Port))
	}

	// Load templates
	baseLocalesPath := cfg.Locale.Path

	if _, err := os.Stat(baseLocalesPath); !os.IsNotExist(err) {
		var err error
		I18n, err = i18n.New(i18n.Glob(baseLocalesPath+"/*/*"), cfg.Locale.Langs...)
		if err != nil {
			fmt.Printf("Warning no locale dir: %s\n", err.Error())
		}
	} else {
		I18n, _ = i18n.New(i18n.Glob(""), "en")
	}

	translateLangFunc := func(x string) string { return Tr(cfg.Locale.Lang, x) }

	r.SetFuncMap(template.FuncMap{
		"tr": translateLangFunc,
	})

	//	r.LoadHTMLGlob(basePath + "/web/templates/**/*.tmpl")
	t := []string{}
	dirs, _ := templateFs.ReadDir("web/templates")
	for k := range dirs {
		files, _ := templateFs.ReadDir("web/templates/" + dirs[k].Name())
		for f := range files {
			t = append(t, fmt.Sprintf("web/templates/%s/%s", dirs[k].Name(), files[f].Name()))
		}
	}
	templ := template.Must(template.New("").Funcs(template.FuncMap{"tr": translateLangFunc}).ParseFS(templateFs, t...))
	r.SetHTMLTemplate(templ)
	if cfg.Debug {
		fmt.Printf("\nTemplates loaded:\n\t- %s\n", strings.Join(t, "\n\t- "))
	}

	// Load static files
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/index.tmpl", gin.H{"appname": cfg.AppName, "appdesc": cfg.AppDesc})
	})

	// Cache static files
	r.Use(setCacheHeaders())

	r.Use(static.Serve("/", EmbedFolder(server, "web/assets")))
	r.Static("/css", "/assets/css")
	// r.Static("/fonts", basePath+"/assets/fonts")
	r.Static("/js", "/assets/js")

	r.Use(setConfig(cfg))

	return r
}

func SetRoutes(cfg *config.WebConfig) *gin.Engine {
	r := initServer(cfg)

	mw := csrf.Middleware(csrf.Options{
		Secret: cfg.Sec.CSRFrandom,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	})

	l := r.Group("auth")
	l.GET("/login", LoginHandlerForm)
	l.POST("/login", LoginHandler)
	l.GET("/logout", LogoutHandler)

	u := r.Group("auth/user")
	u.Use(mw)
	u.Use(Auth("self"))
	u.GET("/:id", UserProfile)
	u.POST("/:id", UserChgPasswd)
	u.POST("/otp/:id", UserChgOTP)
	u.POST("/passapp/:id", UserPassApp)

	admin := r.Group("auth/crud")
	admin.Use(mw)
	admin.Use(Auth("admin"))
	admin.GET("/user/", UserList)
	admin.GET("/user/:id", UserEdit)
	admin.POST("/user/:id", UserUpdate) // for HTML 1.1 form don't have PUT/DELETE methods
	// admin.PUT("/:id", UserUpdate)
	admin.GET("/user/create", UserAdd)
	admin.POST("/user/create", UserCreate)
	admin.POST("/user/del/:id", UserDel) // for HTML 1.1 form don't have PUT/DELETE methods
	// admin.DELETE("/:id", UserDel)

	admin.GET("/reload", CancelChanges)
	admin.GET("/save", SaveChanges)

	admin.GET("/group/", GroupList)
	admin.GET("/group/:id", GroupEdit)
	admin.POST("/group/:id", GroupUpdate)
	admin.GET("/group/create", GroupAdd)
	admin.POST("/group/create", GroupCreate)
	admin.POST("/group/del/:id", GroupDel)

	return r
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func setCacheHeaders() gin.HandlerFunc {
	data := []byte(time.Now().String())
	etag := fmt.Sprintf("%x", md5.Sum(data)) //nolint:gosec  //only for cache headers
	cacheSince := time.Now().Format(http.TimeFormat)
	cacheUntil := time.Now().AddDate(0, 12, 0).Format(http.TimeFormat)
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/css/") ||
			strings.Contains(c.Request.URL.Path, "/js/") ||
			strings.Contains(c.Request.URL.Path, "favicon") {
			c.Header("Cache-Control", "public, max-age=604800, immutable")
			c.Header("ETag", etag)
			c.Header("Last-Modified", cacheSince)
			c.Header("Expires", cacheUntil)
			c.Next()
			return
		}
	}
}

func Auth(rolectl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("Cfg").(config.WebConfig)
		GIDcanChgPass := cfg.CfgUsers.GIDcanChgPass
		GIDAdmin := cfg.CfgUsers.GIDAdmin
		GIDuseOtp := cfg.CfgUsers.GIDuseOtp
		username, userid := GetUserID(c)
		if username == "" || userid == "" {
			Log.Info(fmt.Sprintf("%s -- NOK denied old or bad cookie", c.ClientIP()))
			c.Redirect(302, "/auth/logout")
			c.Abort()
			return
		}
		id := GetUserKey(userid)
		role := "user"
		// search admin role
		groups := Data.Users[id].OtherGroups
		groups = append(groups, Data.Users[id].PrimaryGroup)
		if contains(groups, GIDAdmin) {
			role = "admin"
			Log.Info(fmt.Sprintf("%s -- [%s] is admin", c.ClientIP(), username))
		}
		// search allow self change password
		c.Set("Csrf", csrf.GetToken(c))
		c.Set("CanChgPass", false)
		if contains(groups, GIDcanChgPass) || contains(groups, GIDAdmin) {
			c.Set("CanChgPass", true)
		}
		c.Set("UseOtp", false)
		if contains(groups, GIDuseOtp) {
			c.Set("UseOtp", true)
		}
		c.Set("Login", username)
		c.Set("LoginID", userid)
		c.Set("Role", role)
		c.Set("AppName", cfg.AppName)
		Log.Info(fmt.Sprintf("%s -- OK [%s] (%s) valid access", c.ClientIP(), username, userid))
	}
}
