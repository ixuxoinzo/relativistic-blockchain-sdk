package types

import "time"

type PositionUpdateRequest struct {
	Position Position `json:"position" binding:"required"`
}

type PropagationRequest struct {
	Source  string   `json:"source" binding:"required"`
	Targets []string `json:"targets" binding:"required"`
}

type InterplanetaryRequest struct {
	PlanetA string `json:"planet_a" binding:"required"`
	PlanetB string `json:"planet_b" binding:"required"`
}

type ValidationRequest struct {
	Timestamp  time.Time `json:"timestamp" binding:"required"`
	Position   Position  `json:"position" binding:"required"`
	OriginNode string    `json:"origin_node" binding:"required"`
}

type BlockValidationRequest struct {
	Block      *Block `json:"block" binding:"required"`
	OriginNode string `json:"origin_node" binding:"required"`
}

type BatchValidationRequest struct {
	Items      []*ValidatableItem `json:"items" binding:"required"`
	OriginNode string             `json:"origin_node" binding:"required"`
}

type ConsensusTimingRequest struct {
	Validators []string `json:"validators" binding:"required"`
}

type OffsetCalculationRequest struct {
	NodeID         string   `json:"node_id" binding:"required"`
	ReferenceNodes []string `json:"reference_nodes" binding:"required"`
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenRefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

type PaginationRequest struct {
	Page    int `form:"page,default=1"`
	PerPage int `form:"per_page,default=20"`
}

type FilterRequest struct {
	Region   string `form:"region"`
	Status   string `form:"status"`
	Provider string `form:"provider"`
}

type SearchRequest struct {
	Query  string `form:"query" binding:"required"`
	Limit  int    `form:"limit,default=10"`
	Offset int    `form:"offset,default=0"`
}

type TimeRangeRequest struct {
	From time.Time `form:"from" binding:"required"`
	To   time.Time `form:"to" binding:"required"`
}

type WebSocketSubscribeRequest struct {
	Channels []string `json:"channels" binding:"required"`
}

type WebSocketUnsubscribeRequest struct {
	Channels []string `json:"channels" binding:"required"`
}

type WebSocketAuthRequest struct {
	Token string `json:"token" binding:"required"`
}
