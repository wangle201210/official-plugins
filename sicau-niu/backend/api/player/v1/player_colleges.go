// player_colleges.go defines the response DTOs for the player-facing college
// dropdown used during identity profile selection.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CollegeOptionsReq is the request for the player college dropdown options.
type CollegeOptionsReq struct {
	g.Meta `path:"/plugins/sicau-niu/colleges" method:"get" tags:"Sicau Niu Player" summary:"List college options" dc:"Return the bounded, sort-ordered college dropdown for player identity selection. The full active set is returned in one response. Requires a valid player token."`
}

// CollegeOptionsRes is the response for the player college dropdown options.
type CollegeOptionsRes struct {
	List []*CollegeOptionItem `json:"list" dc:"College options ordered by sort then ID" eg:"[]"`
}

// CollegeOptionItem defines one college option for player selection.
type CollegeOptionItem struct {
	Id   int64  `json:"id" dc:"College ID" eg:"3"`
	Name string `json:"name" dc:"College name" eg:"信息工程学院"`
}
