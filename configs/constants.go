package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Holding"

	// STATS RULES
	VIEW_CACHE_EXPIRATION = 1 * time.Minute

	// AI RATE LIMIT RULES
	AI_RATE_LIMIT_WINDOW         = 30 * time.Minute
	AI_RATE_LIMIT_MAX_REQUESTS   = 50
	AI_RATE_LIMIT_REQ_PER_MINUTE = 5
	AI_RATE_LIMIT_MAX_TOKENS     = 10_000_000

	// Session Rules
	REFRESH_TOKEN_LENGTH   = 32
	REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
	REFRESH_TOKEN_NAME     = "refresh_token"
	ACCESS_TOKEN_NAME      = "access_token"
	ACCESS_TOKEN_DURATION  = 1 * time.Minute
	JWT_ISSUER             = "holding"

	// JOBS Application Rules
	JOBS_TRACKING_SUBJECT  = "tracking_jwt"
	JOBS_TRACKING_COOKIE   = "tracking_cookie"
	JOBS_TRACKING_DURATION = 30 * 24 * time.Hour
)
