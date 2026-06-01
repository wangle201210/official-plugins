// This file defines shared DTOs for the backend-only UIdentity CAS plugin API.
package v1

// ResourceRecord is a generic record projection for plugin-owned UIdentity tables.
type ResourceRecord map[string]any

// ResourceListRes returns a paged resource list.
type ResourceListRes struct {
	List  []ResourceRecord `json:"list" dc:"Paged resource records with API field names" eg:"[]"`
	Total int              `json:"total" dc:"Total number of records matching the current filters" eg:"20"`
}

// ResourceGetRes returns one resource record.
type ResourceGetRes struct {
	Data ResourceRecord `json:"data" dc:"Resource record with API field names" eg:"{}"`
}

// ResourceCreateRes returns the created record ID.
type ResourceCreateRes struct {
	Id int64 `json:"id" dc:"Created record ID" eg:"1"`
}

// ResourceUpdateRes is an empty update response.
type ResourceUpdateRes struct{}

// ResourceDeleteRes is an empty delete response.
type ResourceDeleteRes struct{}

// StatItem exposes one aggregate statistic item.
type StatItem struct {
	Name  string `json:"name" dc:"Statistic bucket name" eg:"cas"`
	Total int64  `json:"total" dc:"Statistic bucket total" eg:"10"`
}

// StatsPayload contains aggregate identity and authentication statistics.
type StatsPayload struct {
	AccountCount     int64       `json:"accountCount" dc:"Total account count" eg:"100"`
	AuthCount        int64       `json:"authCount" dc:"Total CAS and OAuth authentication count" eg:"300"`
	AppCount         int64       `json:"appCount" dc:"Total application count" eg:"12"`
	UserByContainer  []*StatItem `json:"userByContainer" dc:"Account count grouped by container" eg:"[]"`
	AppByType        []*StatItem `json:"appByType" dc:"Application count grouped by access model" eg:"[]"`
	AuthByType       []*StatItem `json:"authByType" dc:"Authentication count grouped by protocol type" eg:"[]"`
	CasByAccountType []*StatItem `json:"casByAccountType" dc:"CAS login count grouped by account container" eg:"[]"`
	PassLevel        []*StatItem `json:"passLevel" dc:"Account count grouped by password strength level" eg:"[]"`
	LoginType        []*StatItem `json:"loginType" dc:"CAS login count grouped by login type" eg:"[]"`
	LoginApp         []*StatItem `json:"loginApp" dc:"CAS login count grouped by application alias" eg:"[]"`
}
