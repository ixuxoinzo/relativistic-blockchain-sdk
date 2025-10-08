package types
import (
        "time"
)
type Node struct {
        ID       string    `json:"id"`
        Position Position  `json:"position"`
        Address  string    `json:"address"`
        Metadata Metadata  `json:"metadata"`
        IsActive bool      `json:"is_active"`
        LastSeen time.Time `json:"last_seen"`
}
type Position struct {
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
        Altitude  float64 `json:"altitude"`
}
type Metadata struct {
        Region       string   `json:"region"`
        Provider     string   `json:"provider"`
        Version      string   `json:"version"`
        Capabilities []string `json:"capabilities"`
}
type Block struct {
        Hash         string    `json:"hash"`
        Timestamp    time.Time `json:"timestamp"`
        ProposedBy   string    `json:"proposed_by"`
        NodePosition Position  `json:"node_position"`
        Data         []byte    `json:"data"`
}
type Transaction struct {
        Hash         string    `json:"hash"`
        Timestamp    time.Time `json:"timestamp"`
        NodePosition Position  `json:"node_position"`
        Data         []byte    `json:"data"`
}
type ValidationResult struct {
        BlockHash     string        `json:"block_hash"` 
        Valid         bool          `json:"valid"`
        Reason        string        `json:"reason"`
        Confidence    float64       `json:"confidence"`
        ExpectedDelay time.Duration `json:"expected_delay"`
        ActualDiff    time.Duration `json:"actual_diff"`
        Threshold     float64       `json:"threshold"`
        ErrorCode     string        `json:"error_code,omitempty"`
        ValidatedAt   time.Time     `json:"validated_at"`
}
type PropagationResult struct {
        SourceNode       string        `json:"source_node"`
        TargetNode       string        `json:"target_node"`
        TheoreticalDelay time.Duration `json:"theoretical_delay"`
        ActualDelay      time.Duration `json:"actual_delay"`
        Distance         float64       `json:"distance_km"`
        Success          bool          `json:"success"`
        Timestamp        time.Time     `json:"timestamp"`
}
type NetworkMetrics struct {
        TotalNodes         int            `json:"total_nodes"`
        ActiveNodes        int            `json:"active_nodes"`
        NetworkCoverage    float64        `json:"network_coverage"`
        AverageDelay       time.Duration  `json:"average_delay"`
        MaxDelay           time.Duration  `json:"max_delay"`
        MinDelay           time.Duration  `json:"min_delay"`
        Regions            map[string]int `json:"regions"`
        EngineCalculations int64          `json:"engine_calculations"`
        EngineValidations  int64          `json:"engine_validations"`
        CacheHits          int64          `json:"cache_hits"`
        CacheMisses        int64          `json:"cache_misses"`
        EngineErrors       int64          `json:"engine_errors"`
        CalculatedAt       time.Time      `json:"calculated_at"`
}
type HealthStatus struct {
        Status     string            `json:"status"`
        Timestamp  time.Time         `json:"timestamp"`
        Version    string            `json:"version"`
        NodeCount  int               `json:"node_count"`
        Uptime     string            `json:"uptime"`
        Components map[string]string `json:"components"`
}
type NodeRegistrationRequest struct {
        ID           string   `json:"id"`
        Position     Position `json:"position"`
        Address      string   `json:"address"`
        Region       string   `json:"region"`
        Provider     string   `json:"provider"`
        Version      string   `json:"version"`
        Capabilities []string `json:"capabilities"`
}
type ValidatableItem struct {
        Type        string       `json:"type"`
        Block       *Block       `json:"block,omitempty"`
        Transaction *Transaction `json:"transaction,omitempty"`
}
type Vote struct {
        BlockHash string    `json:"block_hash"`
        VoterID   string    `json:"voter_id"`
        Timestamp time.Time `json:"timestamp"`
        Signature string    `json:"signature"`
}
type Proposal struct {
        Block      *Block    `json:"block"`
        ProposerID string    `json:"proposer_id"`
        Signature  string    `json:"signature"`
        Timestamp  time.Time `json:"timestamp"`
}
