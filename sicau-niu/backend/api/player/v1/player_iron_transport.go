package v1

import "github.com/gogf/gf/v2/frame/g"

type IronTransportStateReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/iron-transport/state" method:"get" tags:"寻牛小程序" summary:"查询铁牛搬运状态" dc:"Return the current bounded iron-cow transport projection, including visible teams, the requesting player's membership, the latest relevant session and at most 100 trace points. Requires a valid player token."`
}

type CreateTransportTeamReq struct {
	g.Meta     `path:"/plugins/sicau-niu/player/iron-transport/teams" method:"post" tags:"寻牛小程序" summary:"创建铁牛搬运队伍" dc:"Create one visible forming transport team led by the current player. The player must not already belong to an active team and a located iron cow must be available. Reusing requestId returns current state without creating another team. Requires a valid player token."`
	Name       string `json:"name" v:"required|max-length:64" dc:"Team display name" eg:"成都校区搬运队"`
	CampusId   string `json:"campusId" v:"required|in:cd,djy,ya" dc:"Campus ID: cd, djy or ya" eg:"cd"`
	MinMembers int    `json:"minMembers" dc:"Minimum members, defaults to current rule" eg:"3"`
	RequestId  string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-create-1b2a3c4d"`
}

type JoinTransportTeamReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/teams/{id}/join" method:"post" tags:"寻牛小程序" summary:"加入铁牛搬运队伍" dc:"Join one visible forming transport team when it has capacity and the current player does not belong to another active team. Reusing requestId returns current state without adding another membership. Requires a valid player token."`
	Id        int64  `json:"id" v:"required" dc:"Target transport team ID" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-join-1b2a3c4d"`
}

type LeaveTransportTeamReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/teams/{id}/leave" method:"post" tags:"寻牛小程序" summary:"离开铁牛搬运队伍" dc:"Leave one forming transport team. A leader leaving ends the team; members may leave independently. An active transport session must be ended first. Reusing requestId returns current state. Requires a valid player token."`
	Id        int64  `json:"id" v:"required" dc:"Target transport team ID" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-leave-1b2a3c4d"`
}

type StartTransportReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/start" method:"post" tags:"寻牛小程序" summary:"开始铁牛搬运" dc:"Start transport for a ready forming team. Only its leader may start, the minimum member count must be reached and no other session may be active. Reusing requestId returns current state. Requires a valid player token."`
	TeamId    int64  `json:"teamId" v:"required" dc:"Transport team ID to start" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-start-1b2a3c4d"`
}

type HeartbeatTransportReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/heartbeat" method:"post" tags:"寻牛小程序" summary:"提交铁牛搬运心跳" dc:"Append one validated GCJ-02 movement point to the active session for a team the current player belongs to. Implausible steps and idle sessions are rejected. Reusing requestId does not append another point. Requires a valid player token."`
	TeamId    int64   `json:"teamId" v:"required" dc:"Active transport team ID" eg:"12"`
	Lat       float64 `json:"lat" v:"required|between:-90,90" dc:"Current GCJ-02 latitude" eg:"30.7059"`
	Lng       float64 `json:"lng" v:"required|between:-180,180" dc:"Current GCJ-02 longitude" eg:"103.8318"`
	RequestId string  `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-heartbeat-1b2a3c4d"`
}

type EndTransportReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/end" method:"post" tags:"寻牛小程序" summary:"结束铁牛搬运" dc:"End the active transport session and team. Only the team leader may end it. Reusing requestId returns current state without ending another session. Requires a valid player token."`
	TeamId    int64  `json:"teamId" v:"required" dc:"Active transport team ID" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-end-1b2a3c4d"`
}

type IronTransportStateRes struct {
	*IronTransportState `json:",inline" dc:"Current complete iron-cow transport state" eg:"{}"`
}

type CreateTransportTeamRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after creating the team" eg:"{}"`
}

type JoinTransportTeamRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after joining the team" eg:"{}"`
}

type LeaveTransportTeamRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after leaving the team" eg:"{}"`
}

type StartTransportRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after starting the session" eg:"{}"`
}

type HeartbeatTransportRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after accepting the heartbeat" eg:"{}"`
}

type EndTransportRes struct {
	*IronTransportState `json:",inline" dc:"Transport state after ending the session" eg:"{}"`
}

type IronTransportState struct {
	Enabled        bool                  `json:"enabled" dc:"Whether a located iron cow makes transport available" eg:"true"`
	Explanation    string                `json:"explanation" dc:"Short transport rule explanation" eg:"铁牛由小队协作搬运，移动距离和活跃时间由服务端轨迹记录。"`
	IdleTimeoutSec int                   `json:"idleTimeoutSec" dc:"Session idle timeout in seconds" eg:"300"`
	MinTeamSize    int                   `json:"minTeamSize" dc:"Default minimum team member count" eg:"3"`
	CampusId       string                `json:"campusId" dc:"Current view campus ID: cd, djy or ya" eg:"cd"`
	IronCow        *IronTransportCow     `json:"ironCow" dc:"Current iron-cow location and movement projection" eg:"{}"`
	Teams          []*IronTransportTeam  `json:"teams" dc:"Visible forming or active teams, bounded to 20" eg:"[]"`
	MyTeamId       string                `json:"myTeamId,omitempty" dc:"Current player's active team ID; omitted when not in a team" eg:"12"`
	Session        *IronTransportSession `json:"session,omitempty" dc:"Latest relevant active or timed-out session; omitted when none" eg:"null"`
}

type IronTransportCow struct {
	Id          string  `json:"id" dc:"Iron-cow ID" eg:"3"`
	Name        string  `json:"name" dc:"Iron-cow display name" eg:"铁牛号"`
	Lat         float64 `json:"lat" dc:"Current GCJ-02 latitude" eg:"30.7058"`
	Lng         float64 `json:"lng" dc:"Current GCJ-02 longitude" eg:"103.8318"`
	Status      string  `json:"status" dc:"Movement status: waiting, forming, moving or timeout" eg:"moving"`
	MovedMeters float64 `json:"movedMeters" dc:"Current session cumulative movement in meters" eg:"128.5"`
}

type IronTransportTeam struct {
	Id         string                 `json:"id" dc:"Transport team ID" eg:"12"`
	Name       string                 `json:"name" dc:"Team display name" eg:"成都校区搬运队"`
	Code       string                 `json:"code" dc:"Short team join code" eg:"A1B2C3"`
	CampusId   string                 `json:"campusId" dc:"Team campus ID: cd, djy or ya" eg:"cd"`
	LeaderId   string                 `json:"leaderId" dc:"Leader player ID" eg:"8"`
	MinMembers int                    `json:"minMembers" dc:"Minimum members required to start" eg:"3"`
	MaxMembers int                    `json:"maxMembers" dc:"Maximum team member count" eg:"6"`
	Members    []*IronTransportMember `json:"members" dc:"Current active team members" eg:"[]"`
	Visible    bool                   `json:"visible" dc:"Whether the team appears in the join list" eg:"true"`
}

type IronTransportMember struct {
	Id       string `json:"id" dc:"Member player ID" eg:"8"`
	Name     string `json:"name" dc:"Member nickname" eg:"川农同学"`
	Avatar   string `json:"avatar" dc:"Member avatar URL" eg:"https://example.com/avatar.png"`
	Role     string `json:"role" dc:"Member role: leader or member" eg:"leader"`
	JoinedAt *int64 `json:"joinedAt" dc:"Join time as Unix timestamp in milliseconds" eg:"1776333600000"`
}

type IronTransportSession struct {
	Id           string                `json:"id" dc:"Transport session ID" eg:"21"`
	TeamId       string                `json:"teamId" dc:"Owning transport team ID" eg:"12"`
	Status       string                `json:"status" dc:"Session status: active, idle_timeout or ended" eg:"active"`
	StartedAt    *int64                `json:"startedAt" dc:"Start time as Unix timestamp in milliseconds" eg:"1776333600000"`
	LastActiveAt *int64                `json:"lastActiveAt" dc:"Last accepted movement time as Unix timestamp in milliseconds" eg:"1776333660000"`
	MovedMeters  float64               `json:"movedMeters" dc:"Cumulative accepted movement in meters" eg:"128.5"`
	Trace        []*IronTransportPoint `json:"trace" dc:"Recent movement points ordered chronologically and bounded to 100" eg:"[]"`
}

type IronTransportPoint struct {
	Lat float64 `json:"lat" dc:"GCJ-02 latitude" eg:"30.7059"`
	Lng float64 `json:"lng" dc:"GCJ-02 longitude" eg:"103.8318"`
	At  *int64  `json:"at" dc:"Recorded time as Unix timestamp in milliseconds" eg:"1776333660000"`
}
