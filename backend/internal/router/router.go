package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/auth"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/config"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/handlers"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/storage"
)

// New builds the full HTTP router with middleware, repos, and handlers wired.
func New(cfg *config.Config, db *pgxpool.Pool, s3 *storage.Client, jwtMgr *auth.Manager) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// repos
	peopleRepo := repo.NewPeopleRepo(db)
	clientsRepo := repo.NewClientsRepo(db)
	activitiesRepo := repo.NewActivitiesRepo(db)
	regRepo := repo.NewRegulatoryRepo(db)
	stkRepo := repo.NewStakeholdersRepo(db)
	consultRepo := repo.NewConsultationsRepo(db)
	tmplRepo := repo.NewTemplatesRepo(db)
	pubRepo := repo.NewPublicationsRepo(db)

	// handlers
	healthH := handlers.NewHealthHandler(db)
	authH := handlers.NewAuthHandler(db, jwtMgr)
	peopleH := handlers.NewPeopleHandler(peopleRepo)
	clientsH := handlers.NewClientsHandler(clientsRepo)
	activitiesH := handlers.NewActivitiesHandler(activitiesRepo, db, s3)
	regH := handlers.NewRegulatoryHandler(regRepo)
	stkH := handlers.NewStakeholdersHandler(stkRepo)
	consultH := handlers.NewConsultationsHandler(consultRepo)
	tmplH := handlers.NewTemplatesHandler(tmplRepo)
	pubH := handlers.NewPublicationsHandler(pubRepo)
	filesH := handlers.NewFilesHandler(s3)
	membershipH := handlers.NewMembershipHandler(db, tmplRepo)
	membersH := handlers.NewMembersHandler(db)
	hrH := handlers.NewHRHandler(db)
	auditH := handlers.NewAuditHandler(db)
	dashH := handlers.NewDashboardHandler(db)

	r.GET("/healthz", healthH.Check)

	v1 := r.Group("/api/v1")
	{
		// public
		v1.POST("/auth/login", authH.Login)

		// authenticated
		authed := v1.Group("")
		authed.Use(jwtMgr.Middleware(), auth.AuditMiddleware(db))
		{
			authed.GET("/auth/me", authH.Me)
			authed.POST("/auth/verify-pin", jwtMgr.VerifyPin)

			authed.GET("/dashboard/summary", dashH.Summary)

			// people (read for all, write for lead+)
			authed.GET("/people", peopleH.List)
			authed.GET("/people/:id", peopleH.Get)
			authed.POST("/people", auth.RequireRole("admin", "lead", "hr"), peopleH.Create)
			authed.PATCH("/people/:id", auth.RequireRole("admin", "lead", "hr"), peopleH.Patch)
			authed.DELETE("/people/:id", auth.RequireRole("admin"), peopleH.Delete)

			// HR-gated payroll/reviews
			hr := authed.Group("/hr")
			hr.Use(auth.RequireRole("admin", "hr"), auth.RequireHRGate())
			{
				hr.GET("/people/:id/salary", peopleH.Get) // real impl would redact in the default path
				hr.GET("/holidays", hrH.ListHolidays)
				hr.POST("/holidays", hrH.CreateHoliday)
				hr.PATCH("/holidays/:id", hrH.PatchHoliday)
				hr.DELETE("/holidays/:id", hrH.DeleteHoliday)
				hr.GET("/holidays/balances", hrH.HolidayBalances)
				hr.GET("/reviews", hrH.ListReviews)
				hr.POST("/reviews", hrH.CreateReview)
				hr.GET("/expenses", hrH.ListExpenses)
				hr.POST("/expenses", hrH.CreateExpense)
				hr.PATCH("/expenses/:id", hrH.PatchExpense)
			}

			authed.GET("/activities", activitiesH.List)
			authed.GET("/activities/:id", activitiesH.Get)
			authed.POST("/activities", auth.RequireRole("admin", "lead", "staff"), activitiesH.Create)
			authed.POST("/activities/:id/clients", auth.RequireRole("admin", "lead", "staff"), activitiesH.LinkClient)
			authed.GET("/activities/:id/outputs", activitiesH.ListOutputs)
			authed.POST("/activities/:id/outputs", auth.RequireRole("admin", "lead", "staff"), activitiesH.UploadOutput)
			authed.GET("/activities/:id/outputs/:fileId/download", activitiesH.DownloadOutput)
			authed.DELETE("/activities/:id", auth.RequireRole("admin", "lead"), activitiesH.Delete)

			authed.GET("/clients", clientsH.List)
			authed.GET("/clients/:id", clientsH.Get)
			authed.POST("/clients", auth.RequireRole("admin", "lead"), clientsH.Create)
			authed.DELETE("/clients/:id", auth.RequireRole("admin"), clientsH.Delete)

			authed.GET("/regulatory/jurisdictions", regH.ListJurisdictions)
			authed.GET("/regulatory/jurisdictions/:id", regH.GetJurisdiction)
			authed.GET("/regulatory/countries", regH.ListCountries)
			authed.GET("/regulatory/regions", regH.ListRegions)

			authed.GET("/stakeholders/contacts", stkH.ListContacts)
			authed.GET("/stakeholders/contacts/:id", stkH.GetContact)
			authed.POST("/stakeholders/contacts", auth.RequireRole("admin", "lead", "staff"), stkH.CreateContact)
			authed.DELETE("/stakeholders/contacts/:id", auth.RequireRole("admin", "lead"), stkH.DeleteContact)
			authed.GET("/stakeholders/institutions", stkH.ListInstitutions)
			authed.GET("/stakeholders/tree", stkH.Tree)

			authed.GET("/consultations", consultH.List)
			authed.GET("/consultations/:id", consultH.Get)
			authed.POST("/consultations", auth.RequireRole("admin", "lead", "staff"), consultH.Create)
			authed.DELETE("/consultations/:id", auth.RequireRole("admin", "lead"), consultH.Delete)

			authed.GET("/templates", tmplH.List)
			authed.GET("/templates/:id", tmplH.Get)
			authed.POST("/templates", auth.RequireRole("admin", "lead"), tmplH.Create)
			authed.DELETE("/templates/:id", auth.RequireRole("admin"), tmplH.Delete)

			authed.GET("/publications", pubH.List)
			authed.GET("/publications/:id", pubH.Get)
			authed.POST("/publications", auth.RequireRole("admin", "lead", "staff"), pubH.Create)
			authed.DELETE("/publications/:id", auth.RequireRole("admin"), pubH.Delete)

			authed.POST("/files/presign-put", filesH.PresignPut)
			authed.GET("/files/presign-get", filesH.PresignGet)
			authed.DELETE("/files", auth.RequireRole("admin", "lead"), filesH.Delete)

			authed.POST("/memberships/generate", auth.RequireRole("admin", "lead", "staff"), membershipH.Generate)

			authed.GET("/members", membersH.List)
			authed.GET("/members/:id", membersH.Get)
			authed.GET("/members/:id/intel", membersH.Intel)
			authed.POST("/members", auth.RequireRole("admin", "lead", "staff"), membersH.Create)
			authed.PATCH("/members/:id", auth.RequireRole("admin", "lead", "staff"), membersH.Patch)
			authed.DELETE("/members/:id", auth.RequireRole("admin", "lead"), membersH.Delete)

			authed.POST("/templates/:id/render", auth.RequireRole("admin", "lead", "staff"), tmplH.Render)
			authed.GET("/audit-log", auth.RequireRole("admin"), auditH.List)
		}
	}

	return r
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}
