package v1

import "github.com/gogf/gf/v2/frame/g"

type NiuDetailReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/niu/{id}" method:"get" tags:"寻牛小程序" summary:"查询单牛详情" dc:"Return one currently visible cattle with safe location projection, feed statistics, iron bonus, a random quote and the current player's owned main card. Requires a valid player token."`
	Id     int64 `json:"id" v:"required" dc:"Cattle ID" eg:"1"`
}

type NiuDetailRes struct {
	Niu   *VisibleNiuItem `json:"niu" dc:"Safe cattle projection using the same location policy as the map list" eg:"{}"`
	Quote string          `json:"quote" dc:"Random enabled school-history quote" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	Card  *NiuDetailCard  `json:"card" dc:"Main card when owned by the current player; null otherwise" eg:"null"`
}

type NiuDetailCard struct {
	Id        int64  `json:"id" dc:"Card ID" eg:"7"`
	Category  string `json:"category" dc:"Card category: person, event, research, college or spirit" eg:"spirit"`
	Title     string `json:"title" dc:"Card title" eg:"川农大精神"`
	Content   string `json:"content" dc:"Card content" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	ImagePath string `json:"imagePath" dc:"Optional card image storage path" eg:"sicau-niu/card/7.jpg"`
	Owned     bool   `json:"owned" dc:"Whether the current player owns the card" eg:"true"`
}
