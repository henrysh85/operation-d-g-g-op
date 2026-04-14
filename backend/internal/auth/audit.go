package auth

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// AuditMiddleware writes one row to audit_log per successful mutating request
// (POST / PATCH / PUT / DELETE, status < 400). Reads are not logged to keep the
// table focused on change history.
func AuditMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method != "POST" && method != "PATCH" && method != "PUT" && method != "DELETE" {
			return
		}
		if c.Writer.Status() >= 400 {
			return
		}

		claims, ok := ClaimsFrom(c)
		if !ok {
			return
		}

		entity := entityFromPath(c.FullPath())
		meta := map[string]any{
			"method":  method,
			"path":    c.FullPath(),
			"status":  c.Writer.Status(),
			"params":  c.Params,
			"remote":  c.ClientIP(),
			"user_ag": c.Request.UserAgent(),
		}
		raw, _ := json.Marshal(meta)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := pool.Exec(ctx,
			`INSERT INTO audit_log (actor_id, action, entity, entity_id, metadata)
			 VALUES ($1, $2, $3, NULLIF($4,'')::uuid, $5)`,
			claims.UserID, method+" "+entity, entity, c.Param("id"), raw,
		); err != nil {
			log.Warn().Err(err).Str("entity", entity).Msg("audit write failed")
		}
	}
}

// entityFromPath turns "/api/v1/activities/:id/clients" into "activities".
func entityFromPath(p string) string {
	p = strings.TrimPrefix(p, "/api/v1/")
	if i := strings.Index(p, "/"); i >= 0 {
		p = p[:i]
	}
	return p
}
